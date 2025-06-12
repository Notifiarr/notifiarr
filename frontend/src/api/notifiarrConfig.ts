/* Auto-generated. DO NOT EDIT. Generator: https://golift.io/goty
 * Edit the source code and run goty again to make updates.
 */

/**
 * A Weekday specifies a day of the week (Sunday = 0, ...).
 * Copied from stdlib to avoid the String method.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler.Weekday>
 */
export enum Weekday {
  Sunday    = 0,
  Monday    = 1,
  Tuesday   = 2,
  Wednesday = 3,
  Thursday  = 4,
  Friday    = 5,
  Saturday  = 6,
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/configfile.AuthType>
 */
export enum AuthType {
  password = 0,
  header   = 1,
  noauth   = 2,
};

/**
 * Frequency sets the base "how-often" a CronJob is executed.
 * See the Frequency constants.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler.Frequency>
 */
export enum Frequency {
  DeadCron = 0,
  Minutely = 1,
  Hourly   = 2,
  Daily    = 3,
  Weekly   = 4,
  Monthly  = 5,
};

/**
 * Profile is the data returned by the profile GET endpoint.
 * Basically everything.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/client.Profile>
 */
export interface Profile {
  username: string;
  config: Config;
  clientInfo?: ClientInfo;
  isWindows: boolean;
  isLinux: boolean;
  isDarwin: boolean;
  isDocker: boolean;
  isUnstable: boolean;
  isFreeBsd: boolean;
  isSynology: boolean;
  headers?: Record<string, null | string[]>;
  fortune: string;
  upstreamIp: string;
  upstreamAllowed: boolean;
  upstreamHeader: string;
  upstreamType: AuthType;
  languages?: Record<string, null | Record<string, LocalizedLanguage>>;
  /**
   * LoggedIn is only used by the front end. Backend does not set or use it.
   */
  loggedIn: boolean;
  updated: Date;
  flags?: Flags;
  dynamic: boolean;
  webauth: boolean;
  msg?: string;
  logFileInfo?: LogFileInfos;
  configFileInfo?: LogFileInfos;
  expvar: AllData;
  hostInfo?: InfoStat;
  disks?: Record<string, Partition>;
  proxyAllow: boolean;
  poolStats?: Record<string, null | PoolSize>;
  started: Date;
  cmdList?: CmdconfigConfig[];
  checkResults?: CheckResult[];
  checkRunning: boolean;
  checkDisabled: boolean;
  program: string;
  version: string;
  revision: string;
  branch: string;
  buildUser: string;
  buildDate: string;
  goVersion: string;
  os: string;
  arch: string;
  binary: string;
  environment?: Record<string, string>;
  docker: boolean;
  uid: number;
  gid: number;
  ip: string;
  gateway: string;
  ifName: string;
  netmask: string;
  md5: string;
  activeTunnel: string;
  tunnelPoolStats?: Record<string, null | PoolSize>;
};

/**
 * Config represents the data in our config file.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/configfile.Config>
 */
export interface Config extends LogConfig, AppsConfig {
  hostId: string;
  uiPassword: string;
  bindAddr: string;
  sslCertFile: string;
  sslKeyFile: string;
  upstreams?: string[];
  autoUpdate: string;
  unstableCh: boolean;
  timeout: string;
  retries: number;
  snapshot: SnapshotConfig;
  services: ServicesConfig;
  service?: ServiceConfig[];
  apt: boolean;
  watchFiles?: WatchFile[];
  endpoints?: Endpoint[];
  commands?: Command[];
  version: number;
};

/**
 * Config determines which checks to run, etc.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Config>
 */
export interface SnapshotConfig extends Plugins {
  timeout: string;
  interval: string;
  zfsPools?: string[];
  useSudo: boolean;
  monitorRaid: boolean;
  monitorDrives: boolean;
  monitorSpace: boolean;
  allDrives: boolean;
  quotas: boolean;
  ioTop: number;
  psTop: number;
  myTop: number;
  ipmi: boolean;
  ipmiSudo: boolean;
};

/**
 * Plugins is optional configuration for "plugins".
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Plugins>
 */
export interface Plugins {
  nvidia: NvidiaConfig;
  mysql?: MySQLConfig[];
};

/**
 * NvidiaConfig is our input data.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.NvidiaConfig>
 */
export interface NvidiaConfig {
  smiPath: string;
  busIDs?: string[];
  disabled: boolean;
};

/**
 * MySQLConfig allows us to gather a process list for the snapshot.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.MySQLConfig>
 */
export interface MySQLConfig {
  name: string;
  host: string;
  username: string;
  password: string;
  timeout: string;
  /**
   * Only used by service checks, snapshot interval is used for mysql.
   */
  interval: string;
};

/**
 * Config for this Services plugin comes from a config file.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.Config>
 */
export interface ServicesConfig {
  interval: string;
  parallel: number;
  disabled: boolean;
  logFile: string;
};

/**
 * ServiceConfig is a thing we check and report results for.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.ServiceConfig>
 */
export interface ServiceConfig {
  name: string;
  type: string;
  value: string;
  expect: string;
  timeout: string;
  interval: string;
  tags?: Record<string, null | any>;
};

/**
 * WatchFile is the input data needed to watch files.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/filewatch.WatchFile>
 */
export interface WatchFile {
  path: string;
  regex: string;
  skip: string;
  poll: boolean;
  pipe: boolean;
  mustExist: boolean;
  logMatch: boolean;
  disabled: boolean;
};

/**
 * Endpoint contains the cronjob definition and url query parameters.
 * This is the input data to poll a url on a frequency.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig.Endpoint>
 */
export interface Endpoint extends CronJob {
  query?: Record<string, null | string[]>;
  header?: Record<string, null | string[]>;
  template: string;
  name: string;
  url: string;
  method: string;
  body: string;
  follow: boolean;
  validSsl: boolean;
  timeout: string;
};

/**
 * CronJob defines when a job should run.
 * When Frequency is set to:
 * 0 `DeadCron` disables the schedule.
 * 1 `Minutely` uses Seconds.
 * 2 `Hourly` uses Minutes and Seconds.
 * 3 `Daily` uses Hours, Minutes and Seconds.
 * 4 `Weekly` uses DaysOfWeek, Hours, Minutes and Seconds.
 * 5 `Monthly` uses DaysOfMonth, Hours, Minutes and Seconds.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler.CronJob>
 */
export interface CronJob {
  /**
   * Frequency to configure the job. Pass 0 disable the cron.
   */
  frequency: Frequency;
  /**
   * Interval for Daily, Weekly and Monthly Frequencies. 1 = every day/week/month, 2 = every other, and so on.
   */
  interval: number;
  /**
   * AtTimes is a list of 'hours, minutes, seconds' to schedule for Daily/Weekly/Monthly frequencies.
   * Also used in Minutely and Hourly schedules, a bit awkwardly.
   */
  atTimes?: number[][];
  /**
   * DaysOfWeek is a list of days to schedule. 0-6. 0 = Sunday.
   */
  daysOfWeek?: Weekday[];
  /**
   * DaysOfMonth is a list of days to schedule. 1 to 31 or -31 to -1 to count backward.
   */
  daysOfMonth?: number[];
  /**
   * Months to schedule. 1 to 12. 1 = January.
   */
  months?: number[];
};

/**
 * Command contains the input data for a defined command.
 * It also contains some saved data about the command being run.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/commands.Command>
 */
export interface Command extends CmdconfigConfig {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig.Config>
 */
export interface CmdconfigConfig {
  name: string;
  hash: string;
  command?: string;
  shell: boolean;
  log: boolean;
  notify: boolean;
  timeout: string;
  /**
   * Args and ArgValues are not config items. They are calculated on startup.
   */
  args: number;
  argValues?: string[];
};

/**
 * LogConfig allows sending logs to rotating files.
 * Setting an AppName will force log creation even if LogFile and HTTPLog are empty.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/logs.LogConfig>
 */
export interface LogConfig {
  logFile: string;
  debugLog: string;
  httpLog: string;
  logFiles: number;
  logFileMb: number;
  fileMode: number;
  debug: boolean;
  quiet: boolean;
  noUploads: boolean;
};

/**
 * Apps is the input configuration to relay requests to Starr apps.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.AppsConfig>
 */
export interface AppsConfig extends BaseConfig {
  sonarr?: StarrConfig[];
  radarr?: StarrConfig[];
  lidarr?: StarrConfig[];
  readarr?: StarrConfig[];
  prowlarr?: StarrConfig[];
  deluge?: DelugeConfig[];
  qbit?: QbitConfig[];
  rtorrent?: RtorrentConfig[];
  sabnzbd?: SabNZBConfig[];
  nzbget?: NZBGetConfig[];
  transmission?: XmissionConfig[];
  tautulli?: TautulliConfig;
  plex: PlexConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.BaseConfig>
 */
export interface BaseConfig {
  apiKey: string;
  extraKeys?: string[];
  urlbase: string;
  maxBody: number;
  serial: boolean;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.StarrConfig>
 */
export interface StarrConfig extends StarrConfig0, ExtraConfig {};

/**
 * Config is the data needed to poll Radarr or Sonarr or Lidarr or Readarr.
 * At a minimum, provide a URL and API Key.
 * HTTPUser and HTTPPass are used for Basic HTTP auth, if enabled (not common).
 * Username and Password are for non-API paths with native authentication enabled.
 * @see golang: <golift.io/starr.Config>
 */
export interface StarrConfig0 {
  apiKey: string;
  url: string;
  httpPass: string;
  httpUser: string;
  username: string;
  password: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.ExtraConfig>
 */
export interface ExtraConfig {
  name: string;
  timeout: string;
  interval: string;
  validSsl: boolean;
  deletes: number;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.DelugeConfig>
 */
export interface DelugeConfig extends ExtraConfig, DelugeConfig0 {};

/**
 * Config is the data needed to poll Deluge.
 * @see golang: <golift.io/deluge.Config>
 */
export interface DelugeConfig0 {
  url: string;
  password: string;
  httppass: string;
  httpuser: string;
  version: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.QbitConfig>
 */
export interface QbitConfig extends ExtraConfig, QbitConfig0 {};

/**
 * Config is the input data needed to return a Qbit struct.
 * This is setup to allow you to easily pass this data in from a config file.
 * @see golang: <golift.io/qbit.Config>
 */
export interface QbitConfig0 {
  url: string;
  username: string;
  password: string;
  httppass: string;
  httpuser: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.RtorrentConfig>
 */
export interface RtorrentConfig extends ExtraConfig {
  url: string;
  username: string;
  password: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.SabNZBConfig>
 */
export interface SabNZBConfig extends ExtraConfig, SabnzbdConfig {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd.Config>
 */
export interface SabnzbdConfig {
  url: string;
  apiKey: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.NZBGetConfig>
 */
export interface NZBGetConfig extends ExtraConfig, NzbgetConfig {};

/**
 * Config is the input data needed to return a NZBGet struct.
 * This is setup to allow you to easily pass this data in from a config file.
 * @see golang: <golift.io/nzbget.Config>
 */
export interface NzbgetConfig {
  url: string;
  username: string;
  password: string;
};

/**
 * XmissionConfig is the Transmission input configuration.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.XmissionConfig>
 */
export interface XmissionConfig extends ExtraConfig {
  url: string;
  username: string;
  password: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.TautulliConfig>
 */
export interface TautulliConfig extends ExtraConfig, TautulliConfig0 {};

/**
 * Config is the Tautulli configuration.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli.Config>
 */
export interface TautulliConfig0 {
  url: string;
  apiKey: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.PlexConfig>
 */
export interface PlexConfig extends PlexConfig0, ExtraConfig {};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Config>
 */
export interface PlexConfig0 {
  url: string;
  token: string;
};

/**
 * ClientInfo is the client's startup data received from the website.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.ClientInfo>
 */
export interface ClientInfo {
  user: {
    id?: any;
    welcome: string;
    subscriber: boolean;
    patron: boolean;
    devAllowed: boolean;
    dateFormat: PHPDate;
    stopLogs: boolean;
    tunnelUrl: string;
    tunnels?: string[];
    mulery?: MuleryServer[];
  };
  actions: {
    plex: ClientinfoPlexConfig;
    apps: AllAppConfigs;
    dashboard: DashConfig;
    sync: SyncConfig;
    mdblist: MdbListConfig;
    gaps: GapsConfig;
    custom?: CronConfig[];
    snapshot: SnapshotConfig;
  };
  integrityCheck: boolean;
};

/**
 * PHPDate allows us to easily convert a PHP date format in Go.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.PHPDate>
 */
export interface PHPDate {
  php: string;
  fmt: string;
};

/**
 * MuleryServer is data from the website. It's a tunnel's https and wss urls.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.MuleryServer>
 */
export interface MuleryServer {
  tunnel: string;
  socket: string;
  location: string;
};

/**
 * PlexConfig is the website-derived configuration for Plex.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.PlexConfig>
 */
export interface ClientinfoPlexConfig {
  interval: string;
  trackSessions: boolean;
  accountMap: string;
  noActivity: boolean;
  activityDelay: string;
  cooldown: string;
  seriesPc: number;
  moviesPc: number;
};

/**
 * AllAppConfigs is the configuration returned from the notifiarr website for Starr apps.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.AllAppConfigs>
 */
export interface AllAppConfigs {
  lidarr?: AppConfig[];
  prowlarr?: AppConfig[];
  radarr?: AppConfig[];
  readarr?: AppConfig[];
  sonarr?: AppConfig[];
};

/**
 * AppConfig is the data that comes from the website for each Starr app.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.AppConfig>
 */
export interface AppConfig {
  instance: number;
  name: string;
  corrupt: string;
  backup: string;
  interval: string;
  stuck: boolean;
  finished: boolean;
};

/**
 * DashConfig is the configuration returned from the notifiarr website for the dashboard configuration.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.DashConfig>
 */
export interface DashConfig {
  interval: string;
  deluge: boolean;
  lidarr: boolean;
  qbit: boolean;
  radarr: boolean;
  readarr: boolean;
  sonarr: boolean;
  plex: boolean;
  sabnzbd: boolean;
  nzbget: boolean;
  rtorrent: boolean;
  transmission: boolean;
};

/**
 * SyncConfig is the configuration returned from the notifiarr website for CF/RP TraSH sync.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.SyncConfig>
 */
export interface SyncConfig {
  interval: string;
  lidarrInstances?: number[];
  radarrInstances?: number[];
  sonarrInstances?: number[];
  lidarrSync?: string[];
  sonarrSync?: string[];
  radarrSync?: string[];
};

/**
 * MdbListConfig contains the instances we send libraries for, and the interval we do it in.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.MdbListConfig>
 */
export interface MdbListConfig {
  interval: string;
  radarr?: number[];
  sonarr?: number[];
};

/**
 * GapsConfig is the configuration returned from the notifiarr website for Radarr Collection Gaps.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.GapsConfig>
 */
export interface GapsConfig {
  instances?: number[];
  interval: string;
};

/**
 * CronConfig defines a custom GET timer from the website.
 * Used to offload crons to clients.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.CronConfig>
 */
export interface CronConfig {
  name: string;
  interval: string;
  endpoint: string;
  description: string;
};

/**
 * LocalizedLanguage is a language and its display name localized to itself and another (parent) language.
 * @see golang: <github.com/Notifiarr/notifiarr/frontend.LocalizedLanguage>
 */
export interface LocalizedLanguage {
  /**
   * Lang is the parent language code this language Name is localized to.
   */
  lang: string;
  /**
   * Code is the language code of the language.
   */
  code: string;
  /**
   * Name is the display name of the language localized to the parent (Lang) language.
   */
  name: string;
  /**
   * Self is the display name of the language localized in its own language.
   */
  self: string;
};

/**
 * Flags are our CLI input flags.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/configfile.Flags>
 */
export interface Flags {
  verReq: boolean;
  longVerReq: boolean;
  restart: boolean;
  aptHook: boolean;
  updated: boolean;
  pslist: boolean;
  fortune: boolean;
  write: string;
  reset: boolean;
  curl: string;
  configFile: string;
  extraConf?: string[];
  envPrefix: string;
  headers?: string[];
  staticDif: string;
};

/**
 * LogFileInfos holds metadata about files.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/logs.LogFileInfos>
 */
export interface LogFileInfos {
  dirs?: string[];
  size: number;
  list?: LogFileInfo[];
};

/**
 * LogFileInfo is returned by GetAllLogFilePaths.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/logs.LogFileInfo>
 */
export interface LogFileInfo {
  id: string;
  name: string;
  path: string;
  size: number;
  time: Date;
  mode: number;
  used: boolean;
  user: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/mnd.AllData>
 */
export interface AllData {
  logFiles?: Record<string, null | any>;
  apiHits?: Record<string, null | any>;
  httpRequests?: Record<string, null | any>;
  timerEvents?: Record<string, null | Record<string, null | any>>;
  timerCounts?: Record<string, null | any>;
  website?: Record<string, null | any>;
  serviceChecks?: Record<string, null | Record<string, null | any>>;
  apps?: Record<string, null | Record<string, null | any>>;
  fileWatcher?: Record<string, null | any>;
};

/**
 * A HostInfoStat describes the host status.
 * This is not in the psutil but it useful.
 * @see golang: <github.com/shirou/gopsutil/v4/host.InfoStat>
 */
export interface InfoStat {
  hostname: string;
  uptime: number;
  bootTime: number;
  procs: number;
  os: string;
  platform: string;
  platformFamily: string;
  platformVersion: string;
  kernelVersion: string;
  kernelArch: string;
  virtualizationSystem: string;
  virtualizationRole: string;
  hostId: string;
};

/**
 * Partition is used for ZFS pools as well as normal Disk arrays.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Partition>
 */
export interface Partition {
  name: string;
  total: number;
  free: number;
  used: number;
  fsType?: string;
  readOnly?: boolean;
  opts?: string[];
};

/**
 * PoolSize represent the number of open connections per status.
 * @see golang: <golift.io/mulery/client.PoolSize>
 */
export interface PoolSize {
  disconnects: number;
  connecting: number;
  idle: number;
  running: number;
  total: number;
  lastConn: Date;
  lastTry: Date;
  active: boolean;
};

/**
 * CheckResult represents the status of a service.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.CheckResult>
 */
export interface CheckResult {
  name: string;
  state: number;
  output?: Output;
  type: string;
  time: Date;
  since: Date;
  interval: number;
  metadata?: Record<string, null | any>;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.Output>
 */
export interface Output {};

/**
 * ProfilePost is the data sent to the profile POST endpoint when updating the trust profile.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/client.ProfilePost>
 */
export interface ProfilePost {
  username: string;
  password: string;
  authType: AuthType;
  header: string;
  newPass: string;
  upstreams: string;
};

/**
 * Stats for a command's invocations.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/commands.Stats>
 */
export interface Stats {
  runs: number;
  fails: number;
  output: string;
  last: string;
  lastCmd: string;
  lastTime: Date;
  lastArgs?: string[];
};

// Packages parsed:
//   1. github.com/Notifiarr/notifiarr/frontend
//   2. github.com/Notifiarr/notifiarr/pkg/apps
//   3. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex
//   4. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd
//   5. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli
//   6. github.com/Notifiarr/notifiarr/pkg/client
//   7. github.com/Notifiarr/notifiarr/pkg/configfile
//   8. github.com/Notifiarr/notifiarr/pkg/logs
//   9. github.com/Notifiarr/notifiarr/pkg/mnd
//  10. github.com/Notifiarr/notifiarr/pkg/services
//  11. github.com/Notifiarr/notifiarr/pkg/snapshot
//  12. github.com/Notifiarr/notifiarr/pkg/triggers/commands
//  13. github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig
//  14. github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler
//  15. github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig
//  16. github.com/Notifiarr/notifiarr/pkg/triggers/filewatch
//  17. github.com/Notifiarr/notifiarr/pkg/website/clientinfo
//  18. github.com/shirou/gopsutil/v4/host
//  19. golift.io/deluge
//  20. golift.io/mulery/client
//  21. golift.io/nzbget
//  22. golift.io/qbit
//  23. golift.io/starr
