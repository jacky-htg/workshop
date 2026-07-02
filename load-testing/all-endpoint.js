import { _options, _post_baseline, _post_burstTraffic } from "./scenario.js";
import { configDevicePayload } from "./payload/config-device.js";
import { userBlockedPayload } from "./payload/user-blocked.js";

const BASE_URL = __ENV.BASE_URL || 'https://api-motorkux.astra-motor.co.id';
const TOKEN = __ENV.TOKEN || '';
const URL_CONFIG = `/api/config`;
const URL_CONFIG_DEVICE = `/api/config/device`;
const URL_USER_BLOCKED = `/api/user/blocked`;

export const options = _options;

export function baseline() {
  _post_baseline(`${BASE_URL}${URL_CONFIG}`, TOKEN, URL_CONFIG, {});
  _post_baseline(`${BASE_URL}${URL_CONFIG_DEVICE}`, TOKEN, URL_CONFIG_DEVICE, configDevicePayload);
  _post_baseline(`${BASE_URL}${URL_USER_BLOCKED}`, TOKEN, URL_USER_BLOCKED, userBlockedPayload);
}

export function burstTraffic() {
  _post_burstTraffic(`${BASE_URL}${URL_CONFIG}`, TOKEN, URL_CONFIG, {});
  _post_burstTraffic(`${BASE_URL}${URL_CONFIG_DEVICE}`, TOKEN, URL_CONFIG_DEVICE, configDevicePayload);
  _post_burstTraffic(`${BASE_URL}${URL_USER_BLOCKED}`, TOKEN, URL_USER_BLOCKED, userBlockedPayload);
}
