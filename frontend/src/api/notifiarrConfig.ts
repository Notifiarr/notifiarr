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
  website  = 2,
  noauth   = 3,
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
 * Integrations is the data returned by the UI integrations endpoint.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/client.Integrations>
 */
export interface Integrations {
  snapshot: null | Snapshot;
  snapshotAge: Date;
  plex: null | PMSInfo;
  plexAge: Date;
  sessions: null | Sessions;
  sessionsAge: Date;
  dashboard: null | States;
  dashboardAge: Date;
  tautulliUsers: null | Users;
  tautulliUsersAge: Date;
  tautulli: null | Info;
  tautulliAge: Date;
  lidarr: {
    status: null | (null | SystemStatus)[];
    statusAge: null | Date[];
    queue: null | (null | Queue)[];
    queueAge: null | Date[];
  };
  radarr: {
    status: null | (null | RadarrSystemStatus)[];
    statusAge: null | Date[];
    queue: null | (null | RadarrQueue)[];
    queueAge: null | Date[];
  };
  readarr: {
    status: null | (null | ReadarrSystemStatus)[];
    statusAge: null | Date[];
    queue: null | (null | ReadarrQueue)[];
    queueAge: null | Date[];
  };
  sonarr: {
    status: null | (null | SonarrSystemStatus)[];
    statusAge: null | Date[];
    queue: null | (null | SonarrQueue)[];
    queueAge: null | Date[];
  };
  prowlarr: {
    status: null | (null | ProwlarrSystemStatus)[];
    statusAge: null | Date[];
  };
};

/**
 * Snapshot is the output data sent to Notifiarr.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Snapshot>
 */
export interface Snapshot {
  version: string;
  system: InfoStat & AvgStat & {
    username: string;
    cpuPerc: number;
    memFree: number;
    memUsed: number;
    memTotal: number;
    temperatures?: null | Record<string, number>;
    users: number;
    cpuTime: TimesStat;
  };
  raid?: null | RaidData;
  driveAges?: null | Record<string, number>;
  driveTemps?: null | Record<string, number>;
  driveHealth?: null | Record<string, string>;
  diskUsage?: null | Record<string, null | Partition>;
  quotas?: null | Record<string, null | Partition>;
  zfsPools?: null | Record<string, null | Partition>;
  ioTop?: null | IOTopData;
  ioStat?: null | (null | IoStatDisk)[];
  ioStat2?: null | Record<string, IOCountersStat>;
  processes?: null | (null | Process)[];
  mysql?: null | Record<string, null | MySQLServerData>;
  nvidia?: null | (null | NvidiaOutput)[];
  ipmiSensors: null | (null | IPMISensor)[];
  synology?: null | Synology;
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
  /**
   * number of processes
   */
  procs: number;
  /**
   * ex: freebsd, linux
   */
  os: string;
  /**
   * ex: ubuntu, linuxmint
   */
  platform: string;
  /**
   * ex: debian, rhel
   */
  platformFamily: string;
  /**
   * version of the complete OS
   */
  platformVersion: string;
  /**
   * version of the OS kernel (if available)
   */
  kernelVersion: string;
  /**
   * native cpu architecture queried at runtime, as returned by `uname -m` or empty string in case of error
   */
  kernelArch: string;
  virtualizationSystem: string;
  /**
   * guest or host
   */
  virtualizationRole: string;
  /**
   * ex: uuid
   */
  hostId: string;
};

/**
 * @see golang: <github.com/shirou/gopsutil/v4/load.AvgStat>
 */
export interface AvgStat {
  load1: number;
  load5: number;
  load15: number;
};

/**
 * TimesStat contains the amounts of time the CPU has spent performing different
 * kinds of work. Time units are in seconds. It is based on linux /proc/stat file.
 * @see golang: <github.com/shirou/gopsutil/v4/cpu.TimesStat>
 */
export interface TimesStat {
  cpu: string;
  user: number;
  system: number;
  idle: number;
  nice: number;
  iowait: number;
  irq: number;
  softirq: number;
  steal: number;
  guest: number;
  guestNice: number;
};

/**
 * RaidData contains raid information from mdstat and/or megacli.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.RaidData>
 */
export interface RaidData {
  mdstat?: string;
  megacli?: null | (null | MegaCLI)[];
};

/**
 * MegaCLI represents the megaraid cli output.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.MegaCLI>
 */
export interface MegaCLI {
  drive: string;
  target: string;
  adapter: string;
  data: null | Record<string, string>;
};

/**
 * Partition is used for ZFS pools as well as normal Disk arrays.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Partition>
 */
export interface Partition {
  /**
   * Mount point or display name.
   */
  name: string;
  /**
   * Block device path (e.g. /dev/sda1); empty for ZFS pools.
   */
  device?: string;
  total: number;
  free: number;
  used: number;
  fsType?: string;
  readOnly?: boolean;
  opts?: null | string[];
};

/**
 * IOTopData is the data structure for iotop output.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.IOTopData>
 */
export interface IOTopData {
  totalRead: number;
  totalWrite: number;
  currentRead: number;
  currentWrite: number;
  procs: null | (null | IOTopProc)[];
};

/**
 * IOTopProc is part of IOTopData.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.IOTopProc>
 */
export interface IOTopProc {
  pid: number;
  prio: string;
  user: string;
  diskRead: number;
  diskWrite: number;
  swapIn: number;
  io: number;
  command: string;
};

/**
 * IoStatDisk is part of IoStatData.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.IoStatDisk>
 */
export interface IoStatDisk {
  disk_device: string;
  rs: number;
  ws: number;
  ds: number;
  rkBs: number;
  wkBs: number;
  dkBs: number;
  rrqms: number;
  wrqms: number;
  drqms: number;
  rrqm: number;
  wrqm: number;
  drqm: number;
  r_await: number;
  w_await: number;
  d_await: number;
  rareqsz: number;
  wareqsz: number;
  dareqsz: number;
  aqusz: number;
  util: number;
};

/**
 * @see golang: <github.com/shirou/gopsutil/v4/disk.IOCountersStat>
 */
export interface IOCountersStat {
  readCount: number;
  mergedReadCount: number;
  writeCount: number;
  mergedWriteCount: number;
  readBytes: number;
  writeBytes: number;
  readTime: number;
  writeTime: number;
  iopsInProgress: number;
  ioTime: number;
  weightedIO: number;
  name: string;
  serialNumber: string;
  label: string;
};

/**
 * Process is a PID's basic info.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Process>
 */
export interface Process {
  name: string;
  pid: number;
  memPercent: number;
  cpuPercent: number;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.MySQLServerData>
 */
export interface MySQLServerData {
  name: string;
  processes: null | (null | MySQLProcess)[];
  globalstatus: null | Record<string, null | any>;
};

/**
 * MySQLProcess represents the data returned from SHOW PROCESS LIST.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.MySQLProcess>
 */
export interface MySQLProcess {
  id: number;
  user: string;
  host: string;
  db: any;
  command: string;
  time: number;
  state: string;
  info: any;
  /**
   * mariadb
   */
  progress: number;
};

/**
 * NvidiaOutput is what we send to the website.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.NvidiaOutput>
 */
export interface NvidiaOutput {
  name: string;
  driverVersion: string;
  pState: string;
  vBios: string;
  busId: string;
  temperature: number;
  utiliization: number;
  memTotal: number;
  memFree: number;
};

/**
 * IPMISensor contains the data for one sensor.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.IPMISensor>
 */
export interface IPMISensor {
  name: string;
  value: number;
  unit: string;
  state: string;
};

/**
 * Synology is the data we care about from the config file.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Synology>
 */
export interface Synology {
  /**
   * 254263
   */
  last_admin_login_build: string;
  /**
   * Synology DiskStation
   */
  manager: string;
  /**
   * Synology Inc.
   */
  vender: string;
  /**
   * DS1517+
   */
  upnpmodelname: string;
  /**
   * 6.2.3
   */
  udc_check_state: string;
  ha: null | Record<string, string>;
};

/**
 * PMSInfo is the `/` path on Plex.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.PMSInfo>
 */
export interface PMSInfo {
  allowCameraUpload: boolean;
  allowChannelAccess: boolean;
  allowSharing: boolean;
  allowSync: boolean;
  allowTuners: boolean;
  backgroundProcessing: boolean;
  certificate: boolean;
  companionProxy: boolean;
  countryCode: string;
  diagnostics: string;
  Directory: null | (null | Directory)[];
  eventStream: boolean;
  friendlyName: string;
  hubSearch: boolean;
  itemClusters: boolean;
  livetv: number;
  machineIdentifier: string;
  maxUploadBitrate: number;
  maxUploadBitrateReason: string;
  maxUploadBitrateReasonMessage: string;
  mediaProviders: boolean;
  multiuser: boolean;
  myPlex: boolean;
  myPlexMappingState: string;
  myPlexSigninState: string;
  myPlexSubscription: boolean;
  myPlexUsername: string;
  offlineTranscode: number;
  ownerFeatures: string;
  photoAutoTag: boolean;
  platform: string;
  platformVersion: string;
  pluginHost: boolean;
  pushNotifications: boolean;
  readOnlyLibraries: boolean;
  requestParametersInCookie: boolean;
  size: number;
  streamingBrainABRVersion: number;
  streamingBrainVersion: number;
  sync: boolean;
  transcoderActiveVideoSessions: number;
  transcoderAudio: boolean;
  transcoderLyrics: boolean;
  transcoderPhoto: boolean;
  transcoderSubtitles: boolean;
  transcoderVideo: boolean;
  transcoderVideoBitrates: string;
  transcoderVideoQualities: string;
  transcoderVideoResolutions: string;
  updatedAt: number;
  updater: boolean;
  version: string;
  voiceSearch: boolean;
};

/**
 * Directory is part of the PMSInfo.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Directory>
 */
export interface Directory {
  count: number;
  key: string;
  title: string;
};

/**
 * Sessions is the config input data.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Sessions>
 */
export interface Sessions {
  server: string;
  hostId: string;
  sessions: null | (null | Session)[];
};

/**
 * Session is a Plex json struct.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Session>
 */
export interface Session {
  User: User;
  Player: Player;
  TranscodeSession: Transcode;
  addedAt: number;
  art: string;
  audienceRating: number;
  audienceRatingImage: string;
  contentRating: string;
  duration: number;
  guid: string;
  grandparentArt: string;
  grandparentGuid: string;
  grandparentKey: string;
  grandparentRatingKey: string;
  grandparentTheme: string;
  grandparentThumb: string;
  grandparentTitle: string;
  index: number;
  key: string;
  lastViewedAt: number;
  librarySectionID: string;
  librarySectionKey: string;
  librarySectionTitle: string;
  originallyAvailableAt: string;
  parentGuid: string;
  parentIndex: number;
  parentKey: string;
  parentRatingKey: string;
  parentThumb: string;
  parentTitle: string;
  primaryExtraKey: string;
  rating: number;
  ratingImage: string;
  ratingKey: string;
  sessionKey: string;
  studio: string;
  summary: string;
  thumb: string;
  title: string;
  titleSort: string;
  type: string;
  updatedAt: number;
  viewCount: number;
  viewOffset: number;
  year: number;
  Session: {
    bandwidth: number;
    id: string;
    location: string;
  };
  Guid?: null | (null | GUID)[];
  Media?: null | (null | Media)[];
  Rating?: null | (null | Rating)[];
};

/**
 * User is part of a Plex Session.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.User>
 */
export interface User {
  id: string;
  thumb: string;
  title: string;
};

/**
 * Player is part of a Plex Session.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Player>
 */
export interface Player {
  address: string;
  device: string;
  machineIdentifier: string;
  model: string;
  platform: string;
  platformVersion: string;
  product: string;
  profile: string;
  remotePublicAddress: string;
  state: string;
  /**
   * this is not a plex item. We calculate this.
   */
  stateTime: any;
  title: string;
  userID: number;
  vendor: string;
  version: string;
  relayed: boolean;
  local: boolean;
  secure: boolean;
};

/**
 * Transcode is part of a Plex Session.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Transcode>
 */
export interface Transcode {
  audioChannels: number;
  audioCodec: string;
  audioDecision: string;
  container: string;
  context: string;
  duration: number;
  key: string;
  maxOffsetAvailable: number;
  minOffsetAvailable: number;
  progress: number;
  protocol: string;
  remaining: number;
  size: number;
  sourceAudioCodec: string;
  sourceVideoCodec: string;
  speed: number;
  timeStamp: number;
  videoCodec: string;
  videoDecision: string;
  throttled: boolean;
  complete: boolean;
  transcodeHwFullPipeline: boolean;
  transcodeHwRequested: boolean;
};

/**
 * GUID is a reusable type from the Section library.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.GUID>
 */
export interface GUID {
  id: string;
};

/**
 * Media is part of a Plex Session.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Media>
 */
export interface Media {
  aspectRatio: string;
  audioChannels: number;
  audioCodec: string;
  audioProfile: string;
  bitrate: number;
  container: string;
  duration: number;
  height: number;
  id: string;
  protocol: string;
  optimizedForStreaming: boolean;
  videoCodec: string;
  videoFrameRate: string;
  videoProfile: string;
  videoResolution: string;
  width: number;
  selected: boolean;
  Part: null | (null | MediaPart)[];
};

/**
 * MediaPart is part of a Plex Session.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.MediaPart>
 */
export interface MediaPart {
  audioProfile: string;
  bitrate: number;
  container: string;
  decision: string;
  duration: number;
  file: string;
  height: number;
  id: string;
  indexes: string;
  key: string;
  protocol: string;
  selected: boolean;
  size: number;
  optimizedForStreaming: boolean;
  videoProfile: string;
  width: number;
  Stream: null | (null | MediaStream)[];
};

/**
 * MediaStream is part of a Plex Session.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.MediaStream>
 */
export interface MediaStream {
  audioChannelLayout?: string;
  bitDepth?: number;
  bitrate: number;
  bitrateMode?: string;
  channels?: number;
  chromaLocation?: string;
  chromaSubsampling?: string;
  codec: string;
  codedHeight?: number;
  codedWidth?: number;
  colorPrimaries?: string;
  colorTrc?: string;
  decision: string;
  default?: boolean;
  displayTitle: string;
  extendedDisplayTitle: string;
  frameRate?: number;
  hasScalingMatrix?: boolean;
  height?: number;
  id: string;
  index: number;
  language?: string;
  languageCode?: string;
  level?: number;
  location: string;
  profile: string;
  refFrames?: number;
  samplingRate?: number;
  scanType?: string;
  selected?: boolean;
  streamType: number;
  width?: number;
  languageTag?: string;
};

/**
 * Rating is part of Plex metadata.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.Rating>
 */
export interface Rating {
  image: string;
  value?: any;
  type: string;
};

/**
 * States is our compiled states for the dashboard.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/dashboard.States>
 */
export interface States {
  lidarr: null | (null | State)[];
  radarr: null | (null | State)[];
  readarr: null | (null | State)[];
  sonarr: null | (null | State)[];
  nzbget: null | (null | State)[];
  rtorrent: null | (null | State)[];
  qbit: null | (null | State)[];
  deluge: null | (null | State)[];
  sabnzbd: null | (null | State)[];
  transmission: null | (null | State)[];
  plexSessions?: any;
};

/**
 * State is partially filled out once for each app instance.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/dashboard.State>
 */
export interface State {
  /**
   * Shared
   */
  error: string;
  instance: number;
  missing?: number;
  size: number;
  percent?: number;
  upcoming?: number;
  next?: null | (null | Sortable)[];
  latest?: null | (null | Sortable)[];
  onDisk?: number;
  /**
   * How long it took.
   */
  elapsed: string;
  name: string;
  /**
   * Radarr
   */
  movies?: number;
  /**
   * Sonarr
   */
  shows?: number;
  episodes?: number;
  /**
   * Readarr
   */
  authors?: number;
  books?: number;
  editions?: number;
  /**
   * Lidarr
   */
  artists?: number;
  albums?: number;
  tracks?: number;
  /**
   * Downloader
   */
  downloads?: number;
  uploaded?: number;
  incomplete?: number;
  downloaded?: number;
  uploading?: number;
  downloading?: number;
  seeding?: number;
  paused?: number;
  errors?: number;
  month?: number;
  week?: number;
  day?: number;
};

/**
 * Sortable holds data about any Starr item. Kind of a generic data store.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/dashboard.Sortable>
 */
export interface Sortable {
  name: string;
  subName?: string;
  date: Date;
  season?: number;
  episode?: number;
};

/**
 * Users is the entire get_users API response.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli.Users>
 */
export interface Users {
  response: {
    result: string;
    message: string;
    data: null | TautulliUser[];
  };
};

/**
 * User is the user data from the get_users API call.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli.User>
 */
export interface TautulliUser {
  row_id: number;
  user_id: number;
  username: string;
  friendly_name: string;
  thumb: string;
  email: string;
  server_token: string;
  shared_libraries: null | string[];
  filter_all: string;
  filter_movies: string;
  filter_tv: string;
  filter_music: string;
  filter_photos: string;
  /**
   * 1,0 (bool)
   */
  is_active: number;
  /**
   * 1,0 (bool)
   */
  is_admin: number;
  /**
   * 1,0 (bool)
   */
  is_home_user: number;
  /**
   * 1,0 (bool)
   */
  is_allow_sync: number;
  /**
   * 1,0 (bool)
   */
  is_restricted: number;
  /**
   * 1,0 (bool)
   */
  do_notify: number;
  /**
   * 1,0 (bool)
   */
  keep_history: number;
  /**
   * 1,0 (bool)
   */
  allow_guest: number;
};

/**
 * Info represent the data returned by the get_tautulli_info command.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli.Info>
 */
export interface Info {
  tautulli_install_type: string;
  tautulli_version: string;
  tautulli_branch: string;
  tautulli_commit: string;
  tautulli_platform: string;
  tautulli_platform_release: string;
  tautulli_platform_version: string;
  tautulli_platform_linux_distro: string;
  tautulli_platform_device_name: string;
  tautulli_python_version: string;
};

/**
 * SystemStatus is the /api/v1/system/status endpoint.
 * @see golang: <golift.io/starr/lidarr.SystemStatus>
 */
export interface SystemStatus {
  appData: string;
  appName: string;
  authentication: string;
  branch: string;
  buildTime: Date;
  instanceName: string;
  isAdmin: boolean;
  isDebug: boolean;
  isDocker: boolean;
  isLinux: boolean;
  isNetCore: boolean;
  isOsx: boolean;
  isProduction: boolean;
  isUserInteractive: boolean;
  isWindows: boolean;
  migrationVersion: number;
  mode: string;
  osName: string;
  packageAuthor: string;
  packageUpdateMechanism: string;
  packageVersion: string;
  runtimeName: string;
  runtimeVersion: string;
  sqliteVersion: string;
  startTime: Date;
  startupPath: string;
  urlBase: string;
  version: string;
};

/**
 * Queue is the /api/v1/queue endpoint.
 * @see golang: <golift.io/starr/lidarr.Queue>
 */
export interface Queue {
  page: number;
  pageSize: number;
  sortKey: string;
  sortDirection: string;
  totalRecords: number;
  records: null | (null | QueueRecord)[];
};

/**
 * QueueRecord represents the records returns by the /api/v1/queue endpoint.
 * @see golang: <golift.io/starr/lidarr.QueueRecord>
 */
export interface QueueRecord {
  downloadClientHasPostImportCategory: boolean;
  artistId: number;
  albumId: number;
  quality: null | Quality;
  size: number;
  title: string;
  sizeleft: number;
  timeleft: string;
  estimatedCompletionTime: Date;
  status: string;
  trackedDownloadStatus: string;
  statusMessages: null | (null | StatusMessage)[];
  downloadId: string;
  protocol: string;
  downloadClient: string;
  indexer: string;
  outputPath: string;
  downloadForced: boolean;
  id: number;
  errorMessage: string;
};

/**
 * Quality is a download quality profile attached to a movie, book, track or series.
 * It may contain 1 or more profiles.
 * Sonarr nor Readarr use Name or ID in this struct.
 * @see golang: <golift.io/starr.Quality>
 */
export interface Quality {
  name?: string;
  id?: number;
  quality?: null | BaseQuality;
  items?: null | (null | Quality)[];
  allowed: boolean;
  /**
   * Not sure which app had this....
   */
  revision?: null | QualityRevision;
};

/**
 * BaseQuality is a base quality profile.
 * @see golang: <golift.io/starr.BaseQuality>
 */
export interface BaseQuality {
  id: number;
  name: string;
  source?: string;
  resolution?: number;
  modifier?: string;
};

/**
 * QualityRevision is probably used in Sonarr.
 * @see golang: <golift.io/starr.QualityRevision>
 */
export interface QualityRevision {
  version: number;
  real: number;
  isRepack?: boolean;
};

/**
 * StatusMessage represents the status of the item. All apps use this.
 * @see golang: <golift.io/starr.StatusMessage>
 */
export interface StatusMessage {
  title: string;
  messages: null | string[];
};

/**
 * SystemStatus is the /api/v3/system/status endpoint.
 * @see golang: <golift.io/starr/radarr.SystemStatus>
 */
export interface RadarrSystemStatus {
  appData: string;
  appName: string;
  authentication: string;
  branch: string;
  buildTime: Date;
  databaseType: string;
  databaseVersion: string;
  instanceName: string;
  isAdmin: boolean;
  isDebug: boolean;
  isDocker: boolean;
  isLinux: boolean;
  isNetCore: boolean;
  isOsx: boolean;
  isProduction: boolean;
  isUserInteractive: boolean;
  isWindows: boolean;
  migrationVersion: number;
  mode: string;
  osName: string;
  packageAuthor: string;
  packageUpdateMechanism: string;
  packageVersion: string;
  runtimeName: string;
  runtimeVersion: string;
  startTime: Date;
  startupPath: string;
  urlBase: string;
  version: string;
};

/**
 * Queue is the /api/v3/queue endpoint.
 * @see golang: <golift.io/starr/radarr.Queue>
 */
export interface RadarrQueue {
  page: number;
  pageSize: number;
  sortKey: string;
  sortDirection: string;
  totalRecords: number;
  records: null | (null | RadarrQueueRecord)[];
};

/**
 * QueueRecord is part of the activity Queue.
 * @see golang: <golift.io/starr/radarr.QueueRecord>
 */
export interface RadarrQueueRecord {
  downloadClientHasPostImportCategory: boolean;
  movieId: number;
  languages: null | (null | Value)[];
  quality: null | Quality;
  customFormats: null | (null | CustomFormatOutput)[];
  size: number;
  title: string;
  sizeleft: number;
  timeleft: string;
  estimatedCompletionTime: Date;
  status: string;
  trackedDownloadStatus: string;
  trackedDownloadState: string;
  statusMessages: null | (null | StatusMessage)[];
  downloadId: string;
  protocol: string;
  downloadClient: string;
  indexer: string;
  outputPath: string;
  id: number;
  errorMessage: string;
};

/**
 * Value is generic ID/Name struct applied to a few places.
 * @see golang: <golift.io/starr.Value>
 */
export interface Value {
  id: number;
  name: string;
};

/**
 * CustomFormatOutput is the output from the CustomFormat methods.
 * @see golang: <golift.io/starr/radarr.CustomFormatOutput>
 */
export interface CustomFormatOutput {
  id: number;
  name: string;
  includeCustomFormatWhenRenaming: boolean;
  specifications: null | (null | CustomFormatOutputSpec)[];
};

/**
 * CustomFormatOutputSpec is part of a CustomFormatOutput.
 * @see golang: <golift.io/starr/radarr.CustomFormatOutputSpec>
 */
export interface CustomFormatOutputSpec {
  name: string;
  implementation: string;
  implementationName: string;
  infoLink: string;
  negate: boolean;
  required: boolean;
  fields: null | (null | FieldOutput)[];
};

/**
 * FieldOutput is generic Name/Value struct applied to a few places.
 * @see golang: <golift.io/starr.FieldOutput>
 */
export interface FieldOutput {
  advanced?: boolean;
  order?: number;
  helpLink?: string;
  helpText?: string;
  hidden?: string;
  label?: string;
  name: string;
  selectOptionsProviderAction?: string;
  type?: string;
  privacy: string;
  value?: any;
  selectOptions?: null | (null | SelectOption)[];
};

/**
 * SelectOption is part of Field.
 * @see golang: <golift.io/starr.SelectOption>
 */
export interface SelectOption {
  dividerAfter?: boolean;
  order: number;
  value: number;
  hint: string;
  name: string;
};

/**
 * SystemStatus is the /api/v1/system/status endpoint.
 * @see golang: <golift.io/starr/readarr.SystemStatus>
 */
export interface ReadarrSystemStatus {
  appData: string;
  appName: string;
  authentication: string;
  branch: string;
  buildTime: Date;
  databaseType: string;
  databaseVersion: string;
  instanceName: string;
  isAdmin: boolean;
  isDebug: boolean;
  isDocker: boolean;
  isLinux: boolean;
  isMono: boolean;
  isNetCore: boolean;
  isOsx: boolean;
  isProduction: boolean;
  isUserInteractive: boolean;
  isWindows: boolean;
  migrationVersion: number;
  mode: string;
  osName: string;
  osVersion: string;
  packageAuthor: string;
  packageUpdateMechanism: string;
  packageVersion: string;
  runtimeName: string;
  runtimeVersion: string;
  startTime: Date;
  startupPath: string;
  urlBase: string;
  version: string;
};

/**
 * Queue is the /api/v1/queue endpoint.
 * @see golang: <golift.io/starr/readarr.Queue>
 */
export interface ReadarrQueue {
  page: number;
  pageSize: number;
  sortKey: string;
  sortDirection: string;
  totalRecords: number;
  records: null | (null | ReadarrQueueRecord)[];
};

/**
 * QueueRecord is a book from the queue API path.
 * @see golang: <golift.io/starr/readarr.QueueRecord>
 */
export interface ReadarrQueueRecord {
  downloadClientHasPostImportCategory: boolean;
  authorId: number;
  bookId: number;
  quality: null | Quality;
  size: number;
  title: string;
  sizeleft: number;
  timeleft: string;
  estimatedCompletionTime: Date;
  status: string;
  trackedDownloadStatus?: string;
  trackedDownloadState?: string;
  statusMessages?: null | (null | StatusMessage)[];
  downloadId?: string;
  protocol: string;
  downloadClient?: string;
  indexer: string;
  outputPath?: string;
  downloadForced: boolean;
  id: number;
  errorMessage: string;
};

/**
 * SystemStatus is the /api/v3/system/status endpoint.
 * @see golang: <golift.io/starr/sonarr.SystemStatus>
 */
export interface SonarrSystemStatus {
  appData: string;
  appName: string;
  authentication: string;
  branch: string;
  buildTime: Date;
  instanceName: string;
  isAdmin: boolean;
  isDebug: boolean;
  isLinux: boolean;
  isMono: boolean;
  isMonoRuntime: boolean;
  isOsx: boolean;
  isProduction: boolean;
  isUserInteractive: boolean;
  isWindows: boolean;
  mode: string;
  osName: string;
  osVersion: string;
  packageAuthor: string;
  packageUpdateMechanism: string;
  packageVersion: string;
  runtimeName: string;
  runtimeVersion: string;
  sqliteVersion: string;
  startTime: Date;
  startupPath: string;
  urlBase: string;
  version: string;
};

/**
 * Queue is the /api/v3/queue endpoint.
 * @see golang: <golift.io/starr/sonarr.Queue>
 */
export interface SonarrQueue {
  page: number;
  pageSize: number;
  sortKey: string;
  sortDirection: string;
  totalRecords: number;
  records: null | (null | SonarrQueueRecord)[];
};

/**
 * QueueRecord is part of Queue.
 * @see golang: <golift.io/starr/sonarr.QueueRecord>
 */
export interface SonarrQueueRecord {
  downloadClientHasPostImportCategory: boolean;
  id: number;
  seriesId: number;
  episodeId: number;
  language: null | Value;
  quality: null | Quality;
  size: number;
  title: string;
  sizeleft: number;
  timeleft: string;
  estimatedCompletionTime: Date;
  status: string;
  trackedDownloadStatus: string;
  trackedDownloadState: string;
  statusMessages: null | (null | StatusMessage)[];
  downloadId: string;
  protocol: string;
  downloadClient: string;
  indexer: string;
  outputPath: string;
  errorMessage: string;
};

/**
 * SystemStatus is the /api/v1/system/status endpoint.
 * @see golang: <golift.io/starr/prowlarr.SystemStatus>
 */
export interface ProwlarrSystemStatus {
  appData: string;
  appName: string;
  authentication: string;
  branch: string;
  buildTime: Date;
  databaseType: string;
  databaseVersion: string;
  instanceName: string;
  isAdmin: boolean;
  isDebug: boolean;
  isDocker: boolean;
  isLinux: boolean;
  isMono: boolean;
  isNetCore: boolean;
  isOsx: boolean;
  isProduction: boolean;
  isUserInteractive: boolean;
  isWindows: boolean;
  migrationVersion: number;
  mode: string;
  osName: string;
  osVersion: string;
  packageAuthor: string;
  packageUpdateMechanism: string;
  packageVersion: string;
  runtimeName: string;
  runtimeVersion: string;
  startTime: Date;
  startupPath: string;
  urlBase: string;
  version: string;
};

/**
 * Profile is the data returned by the profile GET endpoint.
 * Basically everything.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/client.Profile>
 */
export interface Profile {
  username: string;
  config: Config;
  clientInfo: null | ClientInfo;
  isWindows: boolean;
  isLinux: boolean;
  isDarwin: boolean;
  isDocker: boolean;
  isUnstable: boolean;
  isFreeBsd: boolean;
  isSynology: boolean;
  headers: null | Record<string, null | string[]>;
  fortune: string;
  upstreamIp: string;
  upstreamAllowed: boolean;
  upstreamHeader: string;
  upstreamType: AuthType;
  languages: null | Record<string, null | Record<string, LocalizedLanguage>>;
  triggers: null | TriggerInfo[];
  timers: null | TriggerInfo[];
  schedules: null | TriggerInfo[];
  siteCrons: null | (null | Timer)[];
  plexInfo: null | PMSInfo;
  plexAge: Date;
  /**
   * LoggedIn is only used by the front end. Backend does not set or use it.
   */
  loggedIn: boolean;
  updated: Date;
  flags: null | Flags;
  dynamic: boolean;
  webauth: boolean;
  msg?: string;
  logFileInfo: null | LogFileInfos;
  configFileInfo: null | LogFileInfos;
  expvar: AllData;
  hostInfo: null | InfoStat;
  disks: null | Record<string, null | Partition>;
  proxyAllow: boolean;
  started: Date;
  cmdList: null | (null | CmdconfigConfig)[];
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
  environment: null | Record<string, string>;
  docker: boolean;
  uid: number;
  gid: number;
  ip: string;
  gateway: string;
  ifName: string;
  netmask: string;
  md5: string;
  activeTunnel: string;
  tunnelPoolStats: null | Record<string, null | PoolSize>;
  apiKeyValid: boolean;
  apiKeyError?: string;
};

/**
 * Config represents the data in our config file.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/configfile.Config>
 */
export interface Config extends LogConfig, AppsConfig {
  hostId: string;
  uiPassword: string;
  bindAddr: string;
  noCompress: boolean;
  sslCertFile: string;
  sslKeyFile: string;
  upstreams: null | string[];
  autoUpdate: string;
  unstableCh: boolean;
  timeout: string;
  retries: number;
  snapshot: SnapshotConfig;
  services: ServicesConfig;
  service: null | ServiceConfig[];
  apt: boolean;
  watchFiles: null | (null | WatchFile)[];
  endpoints: null | (null | Endpoint)[];
  commands: null | (null | Command)[];
  version: number;
};

/**
 * Config determines which checks to run, etc.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Config>
 */
export interface SnapshotConfig extends Plugins {
  /**
   * total run time allowed.
   */
  timeout: string;
  /**
   * how often to send snaps (cron).
   */
  interval: string;
  /**
   * zfs pools to monitor.
   */
  zfsPools: null | string[];
  /**
   * use sudo for smartctl commands.
   */
  useSudo: boolean;
  /**
   * include mdstat and/or megaraid.
   */
  monitorRaid: boolean;
  /**
   * smartctl commands.
   */
  monitorDrives: boolean;
  /**
   * get disk usage.
   */
  monitorSpace: boolean;
  /**
   * usage for all drives?
   */
  allDrives: boolean;
  /**
   * usage for user quotas?
   */
  quotas: boolean;
  /**
   * number of processes to include from ioTop
   */
  ioTop: number;
  /**
   * number of processes to include from top (cpu usage)
   */
  psTop: number;
  /**
   * number of processes to include from mysql servers.
   */
  myTop: number;
  /**
   * get ipmi sensor info.
   */
  ipmi: boolean;
  /**
   * use sudo to get ipmi sensor info.
   */
  ipmiSudo: boolean;
};

/**
 * Plugins is optional configuration for "plugins".
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.Plugins>
 */
export interface Plugins {
  nvidia: NvidiaConfig;
  mysql: null | MySQLConfig[];
};

/**
 * NvidiaConfig is our input data.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/snapshot.NvidiaConfig>
 */
export interface NvidiaConfig {
  smiPath: string;
  busIDs: null | string[];
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
  disabled: boolean;
  logFile: string;
};

/**
 * ServiceConfig is a thing we check and report results for.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.ServiceConfig>
 */
export interface ServiceConfig {
  /**
   * Radarr
   */
  name: string;
  /**
   * http
   */
  type: string;
  /**
   * http://some.url
   */
  value: string;
  /**
   * 200
   */
  expect: string;
  /**
   * 10s
   */
  timeout: string;
  /**
   * 1m
   */
  interval: string;
  /**
   * copied to Metadata.
   */
  tags: null | Record<string, null | any>;
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
  query: null | Record<string, null | string[]>;
  header: null | Record<string, null | string[]>;
  template: string;
  name: string;
  url: string;
  method: string;
  body: string;
  /**
   * redirects
   */
  follow: boolean;
  /**
   * https only
   */
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
  atTimes: null | number[][];
  /**
   * DaysOfWeek is a list of days to schedule. 0-6. 0 = Sunday.
   */
  daysOfWeek: null | Weekday[];
  /**
   * DaysOfMonth is a list of days to schedule. 1 to 31 or -31 to -1 to count backward.
   */
  daysOfMonth: null | number[];
  /**
   * Months to schedule. 1 to 12. 1 = January.
   */
  months: null | number[];
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
  argValues: null | string[];
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
  fileMode: string;
  debug: boolean;
  trace: boolean;
  quiet: boolean;
  noUploads: boolean;
};

/**
 * AppsConfig is the input configuration to relay requests to Starr apps.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.AppsConfig>
 */
export interface AppsConfig extends BaseConfig {
  sonarr?: null | StarrConfig[];
  radarr?: null | StarrConfig[];
  lidarr?: null | StarrConfig[];
  readarr?: null | StarrConfig[];
  prowlarr?: null | StarrConfig[];
  deluge?: null | DelugeConfig[];
  qbit?: null | QbitConfig[];
  rtorrent?: null | RtorrentConfig[];
  sabnzbd?: null | SabNZBConfig[];
  nzbget?: null | NZBGetConfig[];
  transmission?: null | XmissionConfig[];
  tautulli: TautulliConfig;
  plex: PlexConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.BaseConfig>
 */
export interface BaseConfig {
  apiKey: string;
  extraKeys: null | string[];
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
  http_pass: string;
  http_user: string;
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
  http_pass: string;
  http_user: string;
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
    username: string;
    email: string;
    hash: string;
    welcome: string;
    subscriber: boolean;
    patron: boolean;
    devAllowed: boolean;
    dateFormat: PHPDate;
    stopLogs: boolean;
    tunnelUrl: string;
    tunnels: null | string[];
    mulery: null | (null | MuleryServer)[];
  };
  actions: Actions;
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
  /**
   * ex: "https://africa.notifiarr.com/"
   */
  tunnel: string;
  /**
   * ex: "wss://africa.notifiarr.com/register"
   */
  socket: string;
  /**
   * ex: "Nairobi, Kenya, Africa"
   */
  location: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.Actions>
 */
export interface Actions {
  /**
   * Site Config for Plex.
   */
  plex: WebsiteConfig;
  /**
   * Site Config for Starr.
   */
  apps: AllAppConfigs;
  /**
   * Site Config for Dashboard.
   */
  dashboard: DashConfig;
  /**
   * Site Config for TRaSH Sync.
   */
  sync: SyncConfig;
  /**
   * Site Config for MDB List.
   */
  mdblist: MdbListConfig;
  /**
   * Site Config for Radarr Gaps.
   */
  gaps: GapsConfig;
  /**
   * Site config for Custom Crons.
   */
  custom: null | (null | CronConfig)[];
  /**
   * Site Config for System Snapshot.
   */
  snapshot: SnapshotConfig;
};

/**
 * WebsiteConfig is the website-derived configuration for Plex.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex.WebsiteConfig>
 */
export interface WebsiteConfig {
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
  lidarr: null | (null | AppConfig)[];
  prowlarr: null | (null | AppConfig)[];
  radarr: null | (null | AppConfig)[];
  readarr: null | (null | AppConfig)[];
  sonarr: null | (null | AppConfig)[];
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
  /**
   * how often to fire.
   */
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
  /**
   * how often to fire.
   */
  interval: string;
  /**
   * which instance IDs we sync
   */
  lidarrInstances: null | number[];
  /**
   * which instance IDs we sync
   */
  radarrInstances: null | number[];
  /**
   * which instance IDs we sync
   */
  sonarrInstances: null | number[];
  /**
   * items in sync.
   */
  lidarrSync: null | string[];
  /**
   * items in sync.
   */
  sonarrSync: null | string[];
  /**
   * items in sync.
   */
  radarrSync: null | string[];
};

/**
 * MdbListConfig contains the instances we send libraries for, and the interval we do it in.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.MdbListConfig>
 */
export interface MdbListConfig {
  /**
   * how often to fire.
   */
  interval: string;
  /**
   * which instance IDs we sync
   */
  radarr: null | number[];
  /**
   * which instance IDs we sync
   */
  sonarr: null | number[];
};

/**
 * GapsConfig is the configuration returned from the notifiarr website for Radarr Collection Gaps.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.GapsConfig>
 */
export interface GapsConfig {
  instances: null | number[];
  interval: string;
};

/**
 * CronConfig defines a custom GET timer from the website.
 * Used to offload crons to clients.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/website/clientinfo.CronConfig>
 */
export interface CronConfig {
  /**
   * name of action.
   */
  name: string;
  /**
   * how often to GET this URI.
   */
  interval: string;
  /**
   * endpoint for the URI.
   */
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
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/common.TriggerInfo>
 */
export interface TriggerInfo {
  name: string;
  key: string;
  interval?: number;
  cron?: null | CronJob;
  runs: number;
  kind: string;
};

/**
 * Timer is used to trigger actions.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/triggers/crontimer.Timer>
 */
export interface Timer extends CronConfig {};

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
  extraConf: null | string[];
  envPrefix: string;
  headers: null | string[];
};

/**
 * LogFileInfos holds metadata about files.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/logs.LogFileInfos>
 */
export interface LogFileInfos {
  dirs: null | string[];
  size: number;
  list: null | (null | LogFileInfo)[];
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
  mode: string;
  used: boolean;
  user: string;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/mnd.AllData>
 */
export interface AllData {
  logFiles: null | Record<string, null | any>;
  apiHits: null | Record<string, null | any>;
  httpRequests: null | Record<string, null | any>;
  timerEvents: null | Record<string, null | Record<string, null | any>>;
  timerCounts: null | Record<string, null | any>;
  website: null | Record<string, null | any>;
  serviceChecks: null | Record<string, null | Record<string, null | any>>;
  apps: null | Record<string, null | Record<string, null | any>>;
  fileWatcher: null | Record<string, null | any>;
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
 * ServicesConfig is the data returned by the services/config GET endpoint.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/client.ServicesConfig>
 */
export interface ClientServicesConfig {
  results: null | (null | CheckResult)[];
  running: boolean;
  disabled: boolean;
};

/**
 * CheckResult represents the status of a service.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/services.CheckResult>
 */
export interface CheckResult {
  /**
   * "Radarr"
   */
  name: string;
  /**
   * 0 = OK, 1 = Warn, 2 = Crit, 3 = Unknown
   */
  state: number;
  /**
   * metadata message must never be nil.
   */
  output: any;
  /**
   * http, tcp, ping
   */
  type: string;
  /**
   * when it was checked, rounded to Microseconds
   */
  time: Date;
  /**
   * how long it has been in this state, rounded to Microseconds
   */
  since: Date;
  /**
   * interval in seconds
   */
  interval: number;
  /**
   * arbitrary info about the service or result.
   */
  metadata: null | Record<string, null | any>;
};

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
  lastArgs: null | string[];
};

/**
 * APIResponse is a standard response to our caller. JSON encoded blobs.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/apps.APIResponse>
 */
export interface APIResponse {
  /**
   * The status always matches the HTTP response.
   */
  status: string;
  /**
   * This message contains the request-specific response payload.
   */
  message?: any;
};

/**
 * CheckAllOutput is the output from a check all instances test.
 * The JSON keys are used for human display, so ya.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/checkapp.CheckAllOutput>
 */
export interface CheckAllOutput {
  Sonarr: null | TestResult[];
  Radarr: null | TestResult[];
  Readarr: null | TestResult[];
  Lidarr: null | TestResult[];
  Prowlarr: null | TestResult[];
  Plex: null | TestResult[];
  Tautulli: null | TestResult[];
  NZBGet: null | TestResult[];
  Deluge: null | TestResult[];
  Qbittorrent: null | TestResult[];
  Rtorrent: null | TestResult[];
  Transmission: null | TestResult[];
  SabNZB: null | TestResult[];
  timeMS: number;
  elapsed: number;
  workers: number;
  instances: number;
};

/**
 * TestResult is the result from an instance test.
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/checkapp.TestResult>
 */
export interface TestResult {
  status: number;
  message: string;
  elapsed: string;
  config: ExtraConfig;
};

/**
 * @see golang: <github.com/Notifiarr/notifiarr/pkg/client.BrowseDir>
 */
export interface BrowseDir {
  /**
   * Filepath separator.
   */
  sep: string;
  /**
   * Current directory path.
   */
  path: string;
  /**
   * Parent directory path.
   */
  mom: string;
  /**
   * Directories in the current directory.
   */
  dirs: null | string[];
  /**
   * Files in the current directory.
   */
  files: null | string[];
  /**
   * Error message.
   */
  error: string;
};

// Packages parsed:
//   1. github.com/Notifiarr/notifiarr/frontend
//   2. github.com/Notifiarr/notifiarr/pkg/apps
//   3. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/plex
//   4. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/sabnzbd
//   5. github.com/Notifiarr/notifiarr/pkg/apps/apppkg/tautulli
//   6. github.com/Notifiarr/notifiarr/pkg/checkapp
//   7. github.com/Notifiarr/notifiarr/pkg/client
//   8. github.com/Notifiarr/notifiarr/pkg/configfile
//   9. github.com/Notifiarr/notifiarr/pkg/logs
//  10. github.com/Notifiarr/notifiarr/pkg/mnd
//  11. github.com/Notifiarr/notifiarr/pkg/services
//  12. github.com/Notifiarr/notifiarr/pkg/snapshot
//  13. github.com/Notifiarr/notifiarr/pkg/triggers/commands
//  14. github.com/Notifiarr/notifiarr/pkg/triggers/commands/cmdconfig
//  15. github.com/Notifiarr/notifiarr/pkg/triggers/common
//  16. github.com/Notifiarr/notifiarr/pkg/triggers/common/scheduler
//  17. github.com/Notifiarr/notifiarr/pkg/triggers/crontimer
//  18. github.com/Notifiarr/notifiarr/pkg/triggers/dashboard
//  19. github.com/Notifiarr/notifiarr/pkg/triggers/endpoints/epconfig
//  20. github.com/Notifiarr/notifiarr/pkg/triggers/filewatch
//  21. github.com/Notifiarr/notifiarr/pkg/website/clientinfo
//  22. github.com/shirou/gopsutil/v4/cpu
//  23. github.com/shirou/gopsutil/v4/disk
//  24. github.com/shirou/gopsutil/v4/host
//  25. github.com/shirou/gopsutil/v4/load
//  26. golift.io/deluge
//  27. golift.io/mulery/client
//  28. golift.io/nzbget
//  29. golift.io/qbit
//  30. golift.io/starr
//  31. golift.io/starr/lidarr
//  32. golift.io/starr/prowlarr
//  33. golift.io/starr/radarr
//  34. golift.io/starr/readarr
//  35. golift.io/starr/sonarr
