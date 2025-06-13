import { faBuildingCircleCheck } from '@fortawesome/sharp-duotone-light-svg-icons'
import type { Config, ServiceConfig } from '../../api/notifiarrConfig'
import { deepCopy } from '../../includes/util'
import { get } from 'svelte/store'
import { profile } from '../../api/profile.svelte'
import { _ } from '../../includes/Translate.svelte'

export const page = {
  id: 'Services',
  i: faBuildingCircleCheck,
  c1: 'steelblue',
  c2: 'coral',
  d1: 'wheat',
  d2: 'blue',
}

export const empty: ServiceConfig = {
  name: '',
  interval: '5m0s',
  timeout: '10s',
  type: 'http',
  value: '',
  expect: '200',
  tags: {},
}

export const merge = (index: number, form: ServiceConfig): Config => {
  const c = deepCopy(get(profile).config)
  if (!c.service) c.service = []
  for (let i = 0; i <= index; i++) {
    if (i === index) c.service[i] = form
    else if (i < c.service.length) c.service[i] = {} as ServiceConfig
  }
  return c
}

export const httpCodes = [
  { label: '100: Continue', value: 100 },
  { label: '101: SwitchingProtocols', value: 101 },
  { label: '102: Processing', value: 102 },
  { label: '103: EarlyHints', value: 103 },
  { label: '200: OK', value: 200 },
  { label: '201: Created', value: 201 },
  { label: '202: Accepted', value: 202 },
  { label: '203: NonAuthoritativeInfo', value: 203 },
  { label: '204: NoContent', value: 204 },
  { label: '205: ResetContent', value: 205 },
  { label: '206: PartialContent', value: 206 },
  { label: '207: MultiStatus', value: 207 },
  { label: '208: AlreadyReported', value: 208 },
  { label: '226: IMUsed', value: 226 },
  { label: '300: MultipleChoices', value: 300 },
  { label: '301: MovedPermanently', value: 301 },
  { label: '302: Found', value: 302 },
  { label: '303: SeeOther', value: 303 },
  { label: '304: NotModified', value: 304 },
  { label: '305: UseProxy', value: 305 },
  { label: '307: TemporaryRedirect', value: 307 },
  { label: '308: PermanentRedirect', value: 308 },
  { label: '400: BadRequest', value: 400 },
  { label: '401: Unauthorized', value: 401 },
  { label: '402: PaymentRequired', value: 402 },
  { label: '403: Forbidden', value: 403 },
  { label: '404: NotFound', value: 404 },
  { label: '405: MethodNotAllowed', value: 405 },
  { label: '406: NotAcceptable', value: 406 },
  { label: '407: ProxyAuthRequired', value: 407 },
  { label: '408: RequestTimeout', value: 408 },
  { label: '411: LengthRequired', value: 411 },
  { label: '412: PreconditionFailed', value: 412 },
  { label: '413: RequestEntityTooLarge', value: 413 },
  { label: '414: RequestURITooLong', value: 414 },
  { label: '415: UnsupportedMediaType', value: 415 },
  { label: '416: RequestedRangeNotSatisfiable', value: 416 },
  { label: '417: ExpectationFailed', value: 417 },
  { label: '418: Teapot', value: 418 },
  { label: '421: MisdirectedRequest', value: 421 },
  { label: '422: UnprocessableEntity', value: 422 },
  { label: '423: Locked', value: 423 },
  { label: '424: FailedDependency', value: 424 },
  { label: '425: TooEarly', value: 425 },
  { label: '426: UpgradeRequired', value: 426 },
  { label: '428: PreconditionRequired', value: 428 },
  { label: '429: TooManyRequests', value: 429 },
  { label: '431: RequestHeaderFieldsTooLarge', value: 431 },
  { label: '451: UnavailableForLegalReasons', value: 451 },
  { label: '500: InternalServerError', value: 500 },
  { label: '501: NotImplemented', value: 501 },
  { label: '502: BadGateway', value: 502 },
  { label: '503: ServiceUnavailable', value: 503 },
  { label: '504: GatewayTimeout', value: 504 },
  { label: '505: HTTPVersionNotSupported', value: 505 },
  { label: '506: VariantAlsoNegotiates', value: 506 },
  { label: '507: InsufficientStorage', value: 507 },
  { label: '508: LoopDetected', value: 508 },
  { label: '510: NotExtended', value: 510 },
  { label: '511: NetworkAuthenticationRequired', value: 511 },
]
