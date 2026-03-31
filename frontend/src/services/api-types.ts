// 基于OpenAPI文档生成的TypeScript类型定义
// 版本: OpenAPI 3.0.0
// 生成时间: 2026-01-11

// 云机操作相关类型

// 云机容器项
export interface AndroidItem {
  id: string; // 云机容器ID
  name: string; // 云机名称
  status: string; // 状态
  indexNum: number; // 云机实例位序号
  dataPath: string; // 云机Data文件在设备里的路径
  modelPath: string; // 云机机型文件在设备里的路径
  image: string; // 云机所用的镜像
  ip: string; // 云机局域网IP
  networkName: string; // 容器网卡名称
  portBindings: Record<string, PortBinding[]>; // 端口绑定
  dns: string; // 云机DNS
  doboxFps: string; // 云机FPS
  doboxWidth: string; // 云机分辨率的宽
  doboxHeight: string; // 云机分辨率的高
  doboxDpi: string; // 云机DPI
  s5User: string; // s5的用户名
  s5Password: string; // s5的密码
  s5IP: string; // s5的IP
  s5Port: string; // s5的端口
  s5Type: string; // 代理类型，0-不开启代理，1-本地域名解析，2-服务器域名解析
  created: string; // 云机容器创建时间
  started: string; // 云机容器上次开启时间
  finished: string; // 云机容器上次关闭时间
}

// 端口绑定
export interface PortBinding {
  HostIp: string;
  HostPort: string;
}

// 获取云机列表请求
export interface GetAndroidContainerListReq {
  name?: string; // 可根据云机名过滤
  running?: boolean; // 根据云机容器是否运行过滤，false是查询所有
  indexNum?: number; // 根据云机实例位序号过滤
}

// 获取云机列表响应
export interface GetAndroidContainerListRes {
  count: number; // 总数
  list: AndroidItem[]; // 安卓云机列表
}

// 创建云机请求
export interface CreateAndroidContainerReq {
  name: string; // 云机名称
  modelId?: string; // 线上机型ID
  modelName?: string; // 线上机型名称
  LocalModel?: string; // 本地机型名称，可不填，填了则使用此机型
  modelStatic?: string; // 本地静态机型名称，可不填，填了则使用此机型
  indexNum?: number; // 实例序号
  imageUrl: string; // 镜像完整地址
  sandboxSize?: string; // 沙盒大小
  dns: string; // 云机DNS
  offset?: string; // 云机的开机时间
  doboxFps?: string; // 云机FPS
  doboxWidth?: string; // 云机分辨率的宽
  doboxHeight?: string; // 云机分辨率的高
  doboxDpi?: string; // 云机DPI
  network?: NetworkConfig; // 独立IP设置
  start?: boolean; // 创建完成开机，默认开机
  mgenable?: string; // magisk开关，0-关，1-开
  gmsenable?: string; // gms开关，0-关，1-开
  latitude?: number; // 纬度
  longitude?: number; // 经度
  countryCode?: string; // 国家代码
  portMappings?: PortMapping[]; // 增加自定义端口映射
  s5User?: string; // s5的用户名
  s5Password?: string; // s5的密码
  s5IP?: string; // s5的IP
  s5Port?: string; // s5的端口
  s5Type?: string; // 代理类型
  mytBridgeName?: string; // myt_bridge网卡名
}

// 网络配置
export interface NetworkConfig {
  gw: string; // 网关
  ip: string; // 云机要设置的IP
  subnet: string; // 掩码
}

// 端口映射
export interface PortMapping {
  containerPort: number; // 容器内端口
  hostPort: number; // 主机端口
  hostIP?: string; // 主机IP
  protocol?: string; // 协议，如 tcp、udp, 默认tcp
}

// 创建云机响应
export interface CreateAndroidContainerRes {
  id: string; // 云机容器ID
}

// 删除云机请求
export interface DeleteAndroidContainerReq {
  name: string; // 云机名称
}

// 删除云机响应
export interface DeleteAndroidContainerRes {}

// 重置云机请求
export interface ResetAndroidContainerReq {
  name: string; // 云机名称
  latitude?: number; // 纬度
  longitude?: number; // 经度
  countryCode?: string; // 国家代码
}

// 重置云机响应
export interface ResetAndroidContainerRes {}

// 切换安卓镜像请求
export interface SwitchAndroidImageReq {
  name: string; // 云机名称
  modelId?: string; // 机型ID
  LocalModel?: string; // 本地机型名称
  modelStatic?: string; // 本地静态机型名称
  imageUrl: string; // 镜像完整地址
  dns: string; // 云机DNS
  offset?: string; // 云机的开机时间
  doboxFps?: string; // 云机FPS
  doboxWidth?: string; // 云机分辨率的宽
  doboxHeight?: string; // 云机分辨率的高
  doboxDpi?: string; // 云机DPI
  network?: NetworkConfig; // 独立IP设置
  start?: boolean; // 创建完成开机，默认不开机
  mgenable?: string; // magisk开关
  gmsenable?: string; // gms开关
  latitude?: number; // 纬度
  longitude?: number; // 经度
  countryCode?: string; // 国家代码
  s5User?: string; // s5的用户名
  s5Password?: string; // s5的密码
  s5IP?: string; // s5的IP
  s5Port?: string; // s5的端口
  s5Type?: string; // 代理类型
}

// 切换安卓镜像响应
export interface SwitchAndroidImageRes {
  message?: string;
}

// 切换机型请求
export interface SwitchAndroidPhoneModelReq {
  name: string; // 云机名称
  modelId?: string; // 机型ID
  localModel?: string;
  modelStatic?: string; // 本地静态机型名称
  latitude?: number; // 纬度
  longitude?: number; // 经度
  countryCode?: string; // 国家代码
}

// 切换机型响应
export interface SwitchAndroidPhoneModelRes {}

// 拉取安卓镜像请求
export interface PullAndroidImageReq {
  imageUrl: string; // 镜像完整地址
}

// 拉取安卓镜像响应
export interface PullAndroidImageRes {}

// 启动安卓请求
export interface StartAndroidContainerReq {
  name: string; // 云机名称
}

// 启动安卓响应
export interface StartAndroidContainerRes {}

// 关闭安卓请求
export interface StopAndroidContainerReq {
  name: string; // 云机名称
}

// 关闭安卓响应
export interface StopAndroidContainerRes {}

// 重启安卓请求
export interface ReStartAndroidContainerReq {
  name: string; // 云机名称
}

// 重启安卓响应
export interface ReStartAndroidContainerRes {}

// 本地镜像项
export interface DockerImageItem {
  id: string; // 镜像ID
  imageUrl: string; // 镜像完整地址
  size: string; // 镜像大小
  created: string; // 创建时间
  labels: Record<string, string>; // 镜像labels
}

// 获取本地镜像列表请求
export interface GetDockerImageListReq {
  imageName?: string; // 根据镜像名过滤
}

// 获取本地镜像列表响应
export interface GetDockerImageListRes {
  count: number; // 总数
  list: DockerImageItem[]; // 本地镜像列表
}

// 删除本地镜像请求
export interface RemoveAndroidImageReq {
  imageUrl?: string; // 要删除的镜像完整地址
}

// 删除本地镜像响应
export interface RemoveAndroidImageRes {}

// 镜像压缩包项
export interface ImageFileItem {
  name: string; // 镜像压缩包名称
  size: string; // 镜像压缩包大小
}

// 获取本地镜像压缩包列表请求
export interface GetAndroidImageTarListReq {
  filename?: string; // 可根据文件名过滤
}

// 获取本地镜像压缩包列表响应
export interface GetAndroidImageTarListRes {
  count: number; // 总数
  list: ImageFileItem[]; // 镜像压缩包列表
}

// 删除本地镜像压缩包请求
export interface RemoveAndroidImageTarReq {
  filename: string; // 要删除的镜像压缩包名称
}

// 删除本地镜像压缩包响应
export interface RemoveAndroidImageTarRes {}

// 导出安卓镜像请求
export interface ExportAndroidImageReq {
  imageUrl: string; // 要导出的镜像完整地址
}

// 导出安卓镜像响应
export interface ExportAndroidImageRes {
  filename: string; // 导出后的镜像包文件名
}

// 下载镜像包请求
export interface DownloadImageTarFileReq {
  filename: string; // 要下载的镜像包名
}

// 下载镜像包响应
export interface DownloadImageTarFileRes {}

// 导入安卓镜像请求
export interface ImportAndroidImageReq {
  file: File; // 导入镜像包文件，tar格式
}

// 导入安卓镜像响应
export interface ImportAndroidImageRes {}

// 导出安卓云机请求
export interface ExportAndroidContainerReq {
  name: string; // 云机名称
}

// 导出安卓云机响应
export interface ExportAndroidContainerRes {
  exportName?: string; // 导出后的云机包文件名
}

// 导入安卓云机请求
export interface ImportAndroidContainerReq {
  file: File; // 导入使用本sdk导出的安卓云机
  indexNum?: number; // 实例序号
  name?: string; // 导入后云机名称
}

// 导入安卓云机响应
export interface ImportAndroidContainerRes {
  name?: string; // 导入后云机名称
}

// 机型项
export interface PhoneModelItem {
  id: string; // 机型ID
  name: string; // 机型名称
  md5: string; // 机型文件MD5
  status: string; // 状态
  currentVersion: number; // 当前版本
  sdk_ver: string; // 对应sdk版本
  createdAt: number; // 创建时间
}

// 获取机型列表请求
export interface GetPhoneModelReq {}

// 获取机型列表响应
export interface GetPhoneModelRes {
  list: PhoneModelItem[]; // 机型列表
  total: number; // 机型数量
}

// 国家代码项
export interface CountryCodeItem {
  countryName: string; // 国家名字
  countryCode: string; // 国家代码
}

// 获取国家代码列表请求
export interface GetCountryCodeListReq {}

// 获取国家代码列表响应
export interface GetCountryCodeListRes {
  count: number; // 总数
  list: CountryCodeItem[]; // 国家代码列表
}

// 设置Macvlan请求
export interface SetMacvlanReq {
  gw: string; // 网关
  subnet: string; // 掩码
}

// 设置Macvlan响应
export interface SetMacvlanRes {}

// 重命名安卓容器请求
export interface RenameAndroidContainerReq {
  name: string; // 云机名称
  newName: string; // 云机新名称
}

// 重命名安卓容器响应
export interface RenameAndroidContainerRes {}

// 机型备份项
export interface PhoneModelBackupItem {
  name: string; // 机型备份名称
}

// 获取机型备份列表请求
export interface GetPhoneModelBackupListReq {}

// 获取机型备份列表响应
export interface GetPhoneModelBackupListRes {
  count: number; // 总数
  list: PhoneModelBackupItem[]; // 机型备份列表
}

// 删除机型备份请求
export interface RemovePhoneModelBackupReq {
  name: string; // 机型备份文件名称
}

// 删除机型备份响应
export interface RemovePhoneModelBackupRes {}

// 保存机型备份请求
export interface SavePhoneModelBackupReq {
  name: string; // 要备份机型数据的云机名称
  suffix: string; // 备份后机型数据的后缀名
}

// 保存机型备份响应
export interface SavePhoneModelBackupRes {}

// 导出机型备份请求
export interface ExportPhoneModelBackupReq {
  name: string; // 备份机型数据文件名称
}

// 导出机型备份响应
export interface ExportPhoneModelBackupRes {}

// 导入机型备份请求
export interface ImportPhoneModelBackupReq {
  file: File; // 导入备份机型数据ZIP包
}

// 导入机型备份响应
export interface ImportPhoneModelBackupRes {}

// 安卓云机执行命令请求
export interface AndroidContainerExecReq {
  name: string; // 云机名称
  command: string[]; // 执行的命令
}

// 安卓云机执行命令响应
export interface AndroidContainerExecRes {}

// 接口认证相关类型

// 设置API认证密码请求
export interface SetApiAuthPasswordReq {
  newPassword: string; // 新密码
  confirmPassword: string; // 确认新密码
}

// 设置API认证密码响应
export interface SetApiAuthPasswordRes {}

// 关闭API认证请求
export interface SetApiAuthCloseReq {}

// 关闭API认证响应
export interface SetApiAuthCloseRes {}

// 云机备份相关类型

// 备份项
export interface BackupItem {
  name: string; // 备份压缩包文件名
  size: string; // 备份压缩包大小
}

// 获取备份列表请求
export interface GetBackupContainerListReq {
  name?: string;
}

// 获取备份列表响应
export interface GetBackupContainerListRes {
  count: number; // 总数
  list: BackupItem[]; // 备份压缩包文件列表
}

// 删除备份请求
export interface RemoveBackupContainerReq {
  name: string; // 备份压缩包文件名
}

// 删除备份响应
export interface RemoveBackupContainerRes {}

// 下载备份请求
export interface DownloadBackupContainerReq {
  name: string; // 备份压缩包文件名
}

// 下载备份响应
export interface DownloadBackupContainerRes {}

// 基本信息相关类型

// 获取API信息请求
export interface GetAPIInfoReq {}

// 获取API信息响应
export interface GetAPIInfoRes {
  latestVersion: number; // 线上最新版本号
  currentVersion: number; // 当前本地版本号
}

// 设备信息项
export interface DeviceInfoItem {
  ip: string; // 网口IP
  ip_1: string; // 网口1的IP
  hwaddr: string; // MAC地址
  hwaddr_1: string; // MAC1地址
  cputemp: number; // CPU温度
  cpuload: string; // CPU负载
  memtotal: string; // 内存总大小
  memuse: string; // 内存已使用大小
  mmctotal: string; // 磁盘总大小
  mmcuse: string; // 磁盘已使用大小
  version: string; // 固件版本
  deviceId: string; // 设备ID
  model: string; // 型号版本
  speed: string; // 网口速率
  mmcread: string; // 磁盘读取量
  mmcwrite: string; // 磁盘写入量
  sysuptime: string; // 设备运行时间
  mmcmodel: string; // 磁盘型号
  mmctemp: string; // 磁盘温度
  network4g: string; // 4G网卡
  netWork_eth0: string; // ETH0网卡
}

// 获取设备信息请求
export interface GetDeviceInfoReq {}

// 获取设备信息响应
export interface GetDeviceInfoRes extends DeviceInfoItem {}

// 终端相关类型

// 连接容器终端请求
export interface LinkContainerExecReq {}

// 连接容器终端响应
export interface LinkContainerExecRes {}

// 连接设备SSH请求
export interface LinkDeviceSSHReq {}

// 连接设备SSH响应
export interface LinkDeviceSSHRes {}

// 修改SSH密码请求
export interface LinkDeviceSSHChangePwdReq {
  username?: string; // 默认user
  password: string; // 新密码
}

// 修改SSH密码响应
export interface LinkDeviceSSHChangePwdRes {}

// 开关SSH root登录请求
export interface LinkDeviceSSHSwitchRootReq {
  enable: boolean; // true-启用root登录，false-禁止root登录
}

// 开关SSH root登录响应
export interface LinkDeviceSSHSwitchRootRes {}

// myt_bridge网卡管理相关类型

// Bridge信息项
export interface BridgeInfo {
  name: string;
  ip: string;
  mask: string;
  cidr: string;
}

// 获取myt_bridge网卡列表请求
export interface GetMytBridgeListReq {}

// 获取myt_bridge网卡列表响应
export interface GetMytBridgeListRes {
  count: number; // 总数
  list: BridgeInfo[]; // myt_bridge网卡列表
}

// 创建myt_bridge网卡请求
export interface CreateMytBridgeReq {
  customName: string; // 自定义名
  cidr: string; // cidr
}

// 创建myt_bridge网卡响应
export interface CreateMytBridgeRes {}

// 删除myt_bridge网卡请求
export interface DeleteMytBridgeReq {
  name: string; // 网卡名
}

// 删除myt_bridge网卡响应
export interface DeleteMytBridgeRes {}

// 更新myt_bridge网卡请求
export interface UpdateMytBridgeReq {
  name: string; // 网卡名
  newCidr: string; // 新cidr
}

// 更新myt_bridge网卡响应
export interface UpdateMytBridgeRes {}

// 本地机型数据管理相关类型

// 本地机型项
export interface LocalModelItem {
  name: string; // 机型文件名称
}

// 获取本地机型列表请求
export interface GetLocalPhoneModelListReq {}

// 获取本地机型列表响应
export interface GetLocalPhoneModelListRes {
  count: number; // 总数
  list: LocalModelItem[]; // 本地机型列表
}

// 删除本地机型请求
export interface RemoveLocalPhoneModelReq {
  name: string; // 机型文件名称
}

// 删除本地机型响应
export interface RemoveLocalPhoneModelRes {}

// 导出本地机型请求
export interface ExportLocalPhoneModelReq {
  name: string; // 机型文件名称
}

// 导出本地机型响应
export interface ExportLocalPhoneModelRes {}

// 导入本地机型请求
export interface ImportLocalPhoneModelReq {
  file: File; // 导入修改后的机型ZIP包
}

// 导入本地机型响应
export interface ImportLocalPhoneModelRes {}

// 服务相关类型

// 重启设备请求
export interface RebootDeviceReq {}

// 重启设备响应
export interface RebootDeviceRes {
  message?: string;
}

// 重置设备数据请求
export interface ResetDeviceDataReq {}

// 重置设备数据响应
export interface ResetDeviceDataRes {}

// 升级服务请求
export interface UpgradeServerReq {}

// 升级服务响应
export interface UpgradeServerRes {
  msg?: string;
}

// 通用响应类型
export interface APIResponse<T = any> {
  code?: number;
  message?: string;
  data?: T;
  error?: string;
}

// 设备信息类型
export interface Device {
  ip: string;
  type: string;
  id: string;
  name: string;
  version: 'v0' | 'v1' | 'v2' | 'v3';
  isOnline: boolean;
  lastSeen: Date;
}
