import { _options, _post_baseline, _post_burstTraffic } from "./scenario.js";
import { configDevicePayload } from "./payload/config-device.js";

const BASE_URL = __ENV.BASE_URL || 'https://api-motorkux.astra-motor.co.id';
const TOKEN = __ENV.TOKEN || '';
const PATH = '/api/config/device'
const URL = `${BASE_URL}${PATH}`;

export const options = _options;

export function baseline() {
  _post_baseline(URL, TOKEN, PATH, configDevicePayload);
}

export function burstTraffic() {
  _post_burstTraffic(URL, TOKEN, PATH, configDevicePayload);
}
