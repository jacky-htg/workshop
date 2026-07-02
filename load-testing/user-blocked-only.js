import { _options, _post_baseline, _post_burstTraffic } from "./scenario.js";
import { userBlockedPayload } from "./payload/user-blocked.js"

const BASE_URL = __ENV.BASE_URL || 'https://api-motorkux.astra-motor.co.id';
const TOKEN = __ENV.TOKEN || '';
const URL = `${BASE_URL}/api/user/blocked`;

export const options = _options;

export function baseline() {
  _post_baseline(URL, TOKEN, userBlockedPayload);
}

export function burstTraffic() {
  _post_burstTraffic(URL, TOKEN, userBlockedPayload);
}
