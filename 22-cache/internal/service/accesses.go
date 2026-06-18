package service

import (
	"context"
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log/slog"
	"strings"
	"workshop/config"
	"workshop/internal/model"
	"workshop/internal/repository"
	"workshop/pkg/app"
	"workshop/pkg/cache"
	"workshop/pkg/errors"

	"github.com/jacky-htg/go-libs/logger"
)

type Accesses interface {
	List(ctx context.Context) (map[int]*model.AccessTree, *errors.BusinessError)
	ScanAccess(ctx context.Context, path string) error
}

type accesses struct {
	db    *sql.DB
	cache cache.CacheClient
	log   logger.Logger
	ttl   config.TTLConfig
	repo  repository.AccessRepository
}

func NewAccesses(db *sql.DB, cache cache.CacheClient, log logger.Logger, ttl config.TTLConfig, repo repository.AccessRepository) Accesses {
	return &accesses{db: db, cache: cache, log: log, ttl: ttl, repo: repo}
}

func (u *accesses) List(ctx context.Context) (map[int]*model.AccessTree, *errors.BusinessError) {
	results := make(map[int]*model.AccessTree)

	key := "accesses:list"
	err := u.cache.GetJSON(ctx, key, &results)
	if err == nil {
		return results, nil
	}

	list, err := u.repo.List(ctx)
	if err != nil {
		return nil, errors.InternalServerErrorWrap(err, "error listing access")
	}

	for _, val := range list {
		if val.ParentID != nil && *val.ParentID == 1 {
			results[val.ID] = &model.AccessTree{
				ID:        val.ID,
				Name:      val.Name,
				Alias:     val.Alias,
				Childrens: []model.Access{},
			}
		} else if val.ParentID != nil {
			if parent, exists := results[*val.ParentID]; exists {
				parent.Childrens = append(parent.Childrens, val)
			}
		}
	}

	if err := u.cache.SetJSONWithExpiry(ctx, key, results, u.ttl.TTLDefault); err != nil {
		u.log.Warn(ctx, "set cache json", slog.Any("error", err))
	}

	return results, nil
}

func (u *accesses) ScanAccess(ctx context.Context, path string) error {
	// Parse route definitions from router file
	routes, err := parseRouteDefinitions(path)
	if err != nil {
		return fmt.Errorf("failed to parse route definitions: %w", err)
	}

	rootID := 1

	mapGroups := make(map[string]*model.Access)
	for _, route := range routes {
		if _, exists := mapGroups[route.Group]; !exists {
			mapGroups[route.Group] = &model.Access{
				ParentID: &rootID,
				Name:     route.Group,
				Alias:    route.Group,
			}
		}
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		u.log.Error(ctx, "error begin tx", slog.Any("error", err))
		return err
	}
	defer tx.Rollback()

	for _, access := range mapGroups {
		err := u.repo.Create(ctx, tx, access)
		if err != nil {
			return err
		}
	}

	list := make([]model.Access, 0)
	for _, route := range routes {
		groupAccess := mapGroups[route.Group]
		if groupAccess == nil {
			return fmt.Errorf("group %s not found for route %s", route.Group, route.Alias)
		}

		list = append(list, model.Access{
			ParentID: &groupAccess.ID,
			Name:     fmt.Sprintf("%s %s", route.Method, route.Path),
			Alias:    route.Alias,
		})
	}

	for _, access := range list {
		err := u.repo.Create(ctx, tx, &access)
		if err != nil {
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	if _, err := u.cache.Del(ctx, []string{"accesses:list"}); err != nil {
		u.log.Error(ctx, "error delete cache", slog.Any("error", err), slog.String("key", "accesses:list"))
	}

	return nil
}

func parseRouteDefinitions(filePath string) ([]app.RouteDefinition, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file: %w", err)
	}

	var routes []app.RouteDefinition

	ast.Inspect(node, func(n ast.Node) bool {
		// Look for composite literal
		compLit, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}

		// Check if it's a slice
		if compLit.Type == nil {
			return true
		}

		// Try to match array type: []app.RouteDefinition
		if arrayType, ok := compLit.Type.(*ast.ArrayType); ok {
			// Get the element type
			if selectorExpr, ok := arrayType.Elt.(*ast.SelectorExpr); ok {
				// Check if it's app.RouteDefinition
				if ident, ok := selectorExpr.X.(*ast.Ident); ok {
					if ident.Name == "app" && selectorExpr.Sel.Name == "RouteDefinition" {
						// Extract each route from the composite literal
						for _, elt := range compLit.Elts {
							if route, err := parseRouteFromCompositeLit(elt); err == nil {
								routes = append(routes, route)
							}
						}
					}
				}
			}
		}

		return true
	})

	if len(routes) == 0 {
		return nil, fmt.Errorf("no route definitions found in %s", filePath)
	}

	return routes, nil
}

func parseRouteFromCompositeLit(expr ast.Expr) (app.RouteDefinition, error) {
	compLit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return app.RouteDefinition{}, fmt.Errorf("not a composite literal")
	}

	route := app.RouteDefinition{}

	for _, elt := range compLit.Elts {
		kv, ok := elt.(*ast.KeyValueExpr)
		if !ok {
			continue
		}

		key, ok := kv.Key.(*ast.Ident)
		if !ok {
			continue
		}

		switch key.Name {
		case "Method":
			if value, ok := kv.Value.(*ast.BasicLit); ok {
				route.Method = strings.Trim(value.Value, `"`)
			}
		case "Path":
			if value, ok := kv.Value.(*ast.BasicLit); ok {
				route.Path = strings.Trim(value.Value, `"`)
			}
		case "Group":
			if value, ok := kv.Value.(*ast.BasicLit); ok {
				route.Group = strings.Trim(value.Value, `"`)
			}
		case "Alias":
			if value, ok := kv.Value.(*ast.BasicLit); ok {
				route.Alias = strings.Trim(value.Value, `"`)
			}
		}
	}

	return route, nil
}
