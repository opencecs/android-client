const fs = require('fs');

const tp = 'src/components/NetworkManagement.vue';
let text = fs.readFileSync(tp, 'utf8');

const anchor = '<el-tab-pane :label="$t(\'network.usageGuide\')" name="help">';
if (text.includes('v-if="$i18n.locale === \'zh-CN\'"') && text.includes('Node Management Guide')) {
    console.log('Already translated!');
    process.exit(0);
}

const start = text.indexOf(anchor);
if (start === -1) throw new Error('anchor not found');

const divStart = text.indexOf('<div', start + anchor.length);
if (divStart === -1) throw new Error('div not found');

let innerCount = 0;
let divEnd = -1;

for (let i = divStart; i < text.length; i++) {
    if (text.substring(i, i+4) === '<div') {
        innerCount++;
    } else if (text.substring(i, i+5) === '</div') {
        innerCount--;
        if (innerCount === 0) {
            divEnd = i + 6;
            break;
        }
    }
}

if (divEnd === -1) throw new Error('div close not found');

let zhBlock = text.substring(divStart, divEnd);

// Wrap the zhBlock in English
let enBlock = zhBlock;

// Translations for enBlock 
const dict = {
    '节点管理说明': 'Node Management Guide',
    '功能概述': 'Feature Overview',
    '网络代理分组功能允许您为虚拟机/容器配置代理节点，实现网络加速和优化。通过配置代理分组，您可以：': 'The proxy group feature allows you to configure proxy nodes for VMs/Containers to achieve network acceleration and optimization. You can:',
    '为不同的虚拟机/容器指定不同的网络代理节点': 'Assign different proxy nodes to different VMs/Containers',
    '实现网络流量的智能路由和负载均衡': 'Enable intelligent routing and load balancing of network traffic',
    '提升特定应用的网络访问速度': 'Improve network access speed for specific applications',
    '支持多种主流代理协议，灵活配置': 'Support multiple mainstream proxy protocols with flexible configuration',
    '配置方式说明': 'Configuration Methods',
    '方式一：订阅地址（推荐）': 'Method 1: Subscription URL (Recommended)',
    '通过输入订阅链接，自动获取和更新多个代理节点。': 'Automatically fetch and update multiple proxy nodes by inputting a subscription link.',
    '优点：': 'Pros:',
    '自动获取多个节点，无需手动配置': 'Automatically fetches multiple nodes without manual configuration',
    '服务商更新节点后自动同步': 'Automatically syncs updates from the service provider',
    '配置简单，只需填入一个订阅链接': 'Simple configuration requiring only one subscription link',
    '适用场景：从第三方服务商购买了代理服务': 'Use Case: Purchased proxy services from third-party providers',
    '方式二：手动添加代理协议': 'Method 2: Manually Add Proxy Protocol',
    '手动输入具体的代理协议配置信息。': 'Manually input specific proxy protocol configuration details.',
    '完全自主控制，可使用自建节点': 'Full control, supporting self-hosted nodes',
    '支持批量添加多个节点': 'Supports adding multiple nodes in bulk',
    '支持8种主流代理协议': 'Supports 8 mainstream proxy protocols',
    '适用场景：拥有自建代理服务器或单独的节点配置': 'Use Case: Utilizing self-hosted proxy servers or individual node configs',
    '订阅地址配置步骤': 'Subscription Config Steps',
    '步骤1：获取订阅地址': 'Step 1: Get Subscription URL',
    '从第三方代理服务商处获取订阅链接。常见的服务商包括：': 'Obtain a subscription link from third-party proxy providers. Common providers include:',
    '各大机场服务商的用户中心': 'User centers of major proxy services',
    'VPN服务提供商的订阅页面': 'Subscription pages of VPN providers',
    '私有代理服务的订阅接口': 'Private proxy service subscription APIs',
    '提示：': 'Tip:',
    '订阅链接通常是以 <code': 'Subscription links typically start with <code',
    '开头的URL地址': 'URL addresses',
    '步骤2：添加到客户端': 'Step 2: Add to Client',
    '点击"新增分组"按钮': 'Click the "Add Group" button',
    '选择类型为"订阅地址"': 'Select "Subscription URL" as the type',
    '输入分组别名（如：香港节点、美国节点等）': 'Enter a group alias (e.g., HK Nodes, US Nodes)',
    '将订阅链接粘贴到"URL地址"输入框': 'Paste the subscription URL into the input field',
    '点击"确定"保存': 'Click "Confirm" to save',
    '步骤3：更新订阅': 'Step 3: Update Subscription',
    '系统会自动定期更新订阅内容，也可以在分组列表中手动点击"刷新"按钮立即更新。': 'The system regularly updates subscriptions automatically, or you can click "Refresh" in the group list to update immediately.',
    '代理协议配置步骤': 'Proxy Protocol Config Steps',
    '选择配置类型：': 'Select Config Type:',
    '点击"新增分组"，选择类型为"代理协议"': 'Click "Add Group", select "Proxy Protocol" type',
    '输入分组别名：': 'Input Group Alias:',
    '为这组节点起一个便于识别的名称': 'Give the node group an easily recognizable name',
    '选择协议类型：': 'Select Protocol Type:',
    '从下拉菜单中选择要配置的协议类型（如vmess、vless、ss等）': 'Choose the protocol from the dropdown (e.g., vmess, vless, ss, etc.)',
    '填写协议信息：': 'Fill Protocol Info:',
    '按照对应的格式填写代理节点配置': 'Fill out the proxy node config according to the required format',
    '点击输入框旁边的"支持的协议格式"可查看详细格式说明': 'Click "Supported Protocol Formats" next to the input box to see detailed formats',
    '批量添加：': 'Bulk Add:',
    '可以使用中英文逗号分隔多个节点配置，一次性添加': 'You can use commas to separate multiple node configs for bulk insertion',
    '保存配置：': 'Save Configuration:',
    '点击"确定"完成添加': 'Click "Confirm" to finish adding',
    '支持的协议格式': 'Supported Protocol Formats',
    '1. VMess协议': '1. VMess Protocol',
    '标准的VMess协议格式，使用base64编码的JSON配置': 'Standard VMess protocol format, using base64 encoded JSON configs',
    '2. VLESS协议': '2. VLESS Protocol',
    '支持WebSocket、gRPC等传输方式': 'Supports WebSocket, gRPC, and other transport methods',
    '3. Shadowsocks协议': '3. Shadowsocks Protocol',
    '遵循SIP002标准格式': 'Complies with the SIP002 standard format',
    '4. Trojan协议': '4. Trojan Protocol',
    '支持TCP、WebSocket等传输方式': 'Supports TCP, WebSocket transport methods',
    '5. SOCKS5代理': '5. SOCKS5 Proxy',
    '用户名和密码可以为空，用斜杠分隔': 'Username and password can be empty, separated by slashes',
    '6. HTTP代理': '6. HTTP Proxy',
    '格式与SOCKS5相同': 'Format is the same as SOCKS5',
    '7. WireGuard协议': '7. WireGuard Protocol',
    '现代化的VPN协议，性能优秀': 'Modernized VPN protocol with excellent performance',
    '8. Hysteria2协议': '8. Hysteria2 Protocol',
    '基于QUIC的高性能代理协议': 'High-performance proxy protocol based on QUIC',
    '常见问题': 'Frequently Asked Questions (FAQ)',
    'Q1: 配置保存失败怎么办？': 'Q1: Config saving failed?',
    '可能的原因：': 'Possible causes:',
    '订阅地址格式不正确（需要是完整的URL）': 'Incorrect subscription URL format (must be full URL)',
    '协议格式不符合规范': 'Protocol format does not comply with specifications',
    '网络连接问题': 'Network connection problems',
    '解决方法：': 'Solution:',
    '检查输入格式，确保网络连接正常': 'Check input format and ensure stable network connection',
    'Q2: 节点无法连接怎么办？': 'Q2: Node cannot connect?',
    '节点服务器不可用或过期': 'Node server is unavailable or expired',
    '配置参数错误': 'Configuration parameter error',
    '本地网络限制': 'Local network limitations',
    '尝试更换其他节点，或联系服务商确认节点状态': 'Try switching to another node, or contact service provider',
    'Q3: 速度较慢怎么办？': 'Q3: Slow connection speed?',
    '建议：': 'Suggestions:',
    '选择地理位置较近的节点': 'Select nodes that are geographically closer',
    '避开高峰时段': 'Avoid peak hours',
    '尝试不同的协议类型': 'Try different protocol types',
    '联系服务商升级套餐': 'Contact provider to upgrade your plan',
    'Q4: 订阅无法更新怎么办？': 'Q4: Subscription cannot update?',
    '订阅链接已失效或过期': 'Subscription link has expired or is invalid',
    '服务商服务器维护中': 'Service provider server under maintenance',
    '联系服务商获取最新的订阅链接': 'Contact the service provider for a new subscription link',
    '重要提示': 'Important Notices',
    '功能互斥说明': 'Mutually Exclusive Features',
    '网络代理分组功能与MacVlan公有网卡功能': 'Network Proxy Group feature and MacVlan Public NIC feature',
    '互相排斥，不能同时使用': 'are mutually exclusive and cannot be used simultaneously',
    '使用MacVlan后，节点管理功能将被禁用': 'Enabling MacVlan will disable the Node Management features',
    '配置节点管理后，建议不要再设置MacVlan': 'After configuring nodes, it is recommended NOT to setup MacVlan',
    '切换功能前请先移除原有配置': 'Remove existing configurations before switching features',
    '使用建议': 'Usage Suggestions',
    '订阅地址更新频率建议设置为每天或每周': 'Recommended to set subscription updates to daily or weekly',
    '定期检查节点可用性，及时更新失效节点': 'Regularly check node availability and update dead nodes',
    '建议配置多个分组，便于不同场景切换': 'It is recommended to configure multiple groups for different scenarios',
    '重要应用建议使用稳定性更好的节点': 'Use more stable nodes for critical applications',
    '安全提醒': 'Security Reminders',
    '请从可信的服务商获取订阅地址': 'Please obtain subscription URLs from trusted providers',
    '不要使用来源不明的免费节点': 'Do NOT use unknown free proxy nodes',
    '定期更换密码和敏感信息': 'Periodically change passwords and sensitive data',
    '注意保护个人隐私和数据安全': 'Always safeguard personal privacy and data security',
    '节点分配说明': 'Node Allocation Guide',
    '节点分配功能用于将已配置好的代理节点指定给具体的云机（容器），实现精细化的网络代理管控。': 'The Node Allocation feature assigns configured proxy nodes to specific VMs (containers) for fine-grained network control.',
    '可以为每台云机独立指定代理节点（指定模式）': 'You can independently assign a node to each VM (Assigned Mode)',
    '也可以在某个分组内随机分配节点（随机模式）': 'You can randomly assign a node from a group (Random Mode)',
    '支持批量为多台云机清除已分配的VPC节点': 'Supports bulk clearing of assigned VPC nodes for multiple VMs',
    '操作步骤': 'Operation Steps',
    '选择设备：': 'Select Device:',
    '在左侧设备列表中点击目标设备': 'Click the target device from the left device list',
    '查看已分配节点：': 'View Assigned Nodes:',
    '右侧面板展示该设备下所有云机的VPC节点分配情况': 'The right panel shows the VPC node allocation for all VMs on this device',
    '分配VPC节点': 'Assign VPC Node',
    '点击"分配VPC"，按照三步向导完成分配': 'Click "Assign VPC" and follow the three-step wizard',
    '选择云机：': 'Select VM:',
    '勾选需要配置的一台或多台云机': 'Check one or more VMs to configure',
    '选择分组：': 'Select Group:',
    '从已有节点分组中选择，并决定是"指定节点"还是"随机节点"': 'Choose from existing node groups and decide if it is "Assigned" or "Random"',
    '选择节点：': 'Select Node:',
    '（指定模式下）选择具体的代理节点后点击"确定"': '(In Assigned mode) select a specific node and click "Confirm"',
    '清除VPC节点：': 'Clear VPC Node:',
    '可单独清除某台云机的节点，或勾选多台后批量清除': 'Clear specific VM nodes or do it in bulk on multiple selections',
    'DNS白名单：': 'DNS Whitelist:',
    '点击"开启DNS"可为该云机启用DNS白名单，再次点击则关闭': 'Click "Enable DNS" to turn on DNS whitelist, click again to disable',
    '注意事项': 'Cautions',
    '设置 MacVlan 后节点管理与节点分配功能将': 'Enabling MacVlan will cause Node Management and Allocation to be',
    '不再可用': 'UNAVAILABLE',
    '分配前请确保已在"节点管理"中创建好代理分组和节点': 'Before allocating, ensure you have created proxy groups and nodes in Node Management',
    '随机模式下每次重启云机可能会使用分组中不同的节点': 'In random mode, VMs may select a different node from the group each time they restart',
    '域名过滤说明': 'Domain Filtering Guide',
    '域名过滤功能可以为云机容器或整个设备设置"域名屏蔽规则"，符合规则的域名请求将被代理': 'Domain Filtering sets domain blocking rules for specific containers or devices. Matching requests are',
    '直接丢弃拦截': 'DROPPED and INTERCEPTED directly',
    '不会转发也不会响应。': 'and will neither be forwarded nor responded to.',
    '容器域名过滤：': 'Container Domain Filtering:',
    '仅对指定的单个云机容器生效': 'Applies ONLY to a specific VM container',
    '全局域名过滤：': 'Global Domain Filtering:',
    '对设备下所有云机容器生效': 'Applies to ALL VM containers on the device',
    '规则': 'Rules',
    '同时支持域名和 IP 地址': 'simultaneously support Domain Names and IPs',
    '两种格式，可在域名过渡期间并行填写新旧域名及对应 IP': ', allowing concurrent inputs of new/old domains alongside IP addresses during transitions',
    '支持三种规则匹配模式：子域名/IP 匹配、完整匹配、关键字匹配': 'Supports 3 match modes: Subdomain/IP, Exact Match, and Keyword',
    '规则匹配类型说明': 'Match Rule Types Guide',
    '子域名 / IP 匹配（推荐）': 'Subdomain / IP Match (Recommended)',
    '匹配该域名及其所有子域名，也支持直接填写 IP 地址。': 'Matches domains and all subdomains, including direct IPs.',
    '例如': 'For example:',
    '会同时匹配': 'will match both',
    '和': 'and',
    '；填写': '; inputting',
    '则直接匹配该 IP': 'will directly block that IP',
    '完整域名 / IP 精确匹配': 'Exact Domain / IP Match',
    '仅精确匹配完整域名或完整 IP。': 'Matches strictly exact domain strings or exact IPs.',
    '仅匹配': 'will ONLY match',
    '，不匹配': ', and NOT match',
    '关键字匹配': 'Keyword Match',
    '匹配包含该关键字的任意域名或 IP。': 'Matches any domain or IP containing the keyword.',
    '匹配所有含 "google" 的域名；': 'will match any domain with "google";',
    '匹配所有含该字段的 IP': 'will match any IP containing that string',
    '设置容器域名过滤': 'Set Container Filter',
    '在左侧选择目标设备': 'Select target device on the left',
    '点击"设置容器域名过滤"按钮': 'Click "Set Container Filter" button',
    '在弹窗中选择目标容器': 'Select the target container in the dialog',
    '添加域名规则（可添加多条），选择匹配类型并填写域名': 'Add domain rules (multiple allowed), choose match type and fill the domain',
    '设置全局域名过滤': 'Set Global Domain Filter',
    '点击"设置全局域名过滤"按钮': 'Click "Set Global Domain Filter" button',
    '添加域名规则后点击"确定"保存': 'Add domain rules and click "Confirm" to save',
    '规则将对该设备下所有容器生效': 'Rules will be applied to ALL containers on the device',
    '查询与清除：': 'Query & Clear:',
    '在容器下拉框中选择容器可查看该容器的过滤规则': 'Select a container in the dropdown to view its filter rules',
    '点击"查询全局过滤"可查看全局规则': 'Click "Query Global Filter" to view global rules',
    '点击"清除过滤"可删除当前查看的规则': 'Click "Clear Filter" to delete the currently viewed rules',
    '域名直连说明': 'Direct Domain Bypass Guide',
    '域名直连功能可以为指定容器设置"直连白名单"，符合规则的域名请求将': 'Direct Domain assigns a "bypass whitelist" to specific containers. Matched domains will',
    '绕过 VPC 代理': 'BYPASS VPC Proxies',
    '，直接走本地网络连接。': ', and connect directly through local networks.',
    '仅对指定容器生效：': 'Applying Only to Specific Containers:',
    '每条规则绑定到单个云机容器': 'Each rule bounds to a single VM container',
    '适用于某些域名需要直接连接、不走代理的场景': 'Ideal for scenarios where certain domains require direct connections instead of proxies',
    '在弹窗中选择目标容器': 'Select target container in the pop-up',
    '添加直连规则（可添加多条），选择匹配类型并填写域名或 IP': 'Add direct rules (multiple allowed), choose match type and fill domain/IP',
    '在容器下拉框中选择容器可查询当前直连规则': 'Select a container in the dropdown to check active direct rules',
    '点击"清除直连"可删除该容器的所有直连规则': 'Click "Clear Bypass" to remove all existing direct rules for the container',
    '私有网卡说明': 'Private NIC Guide',
    '私有网卡（mytBridge）会在设备内创建一个独立的虚拟网桥，为虚拟机/容器分配该网桥下的私有IP地址，从而实现：': 'The Private NIC (mytBridge) creates an isolated virtual bridge on the device to allocate private IP addresses to VMs, enabling:',
    '同一设备上不同云机之间的': 'Traffic between VMs on the same device being',
    '网络隔离': 'NETWORK ISOLATED',
    '仍可使用"节点管理"中配置的IP代理功能': 'Proxy IP features from "Node Management" remain accessible',
    '不影响云机对外访问互联网': 'Does not affect VM access to the external internet',
    '需要多个云机使用不同代理IP，同时防止云机间相互访问': 'Needing distinct VIPs per VM while preventing VMs from accessing each other',
    '创建网卡：': 'Create NIC:',
    '点击"创建网卡"按钮，填写以下信息：': 'Click "Create NIC" and provide the following:',
    '自定义名称：': 'Custom Name:',
    '网卡名称前缀固定为': 'The prefix is permanently set to',
    '，填写后缀部分': ', only fill in the suffix portion',
    '网段地址，例如': 'The subnet CIDR block, e.g.,',
    '，决定虚拟IP的范围': ', strictly defines the IP boundaries',
    '编辑/删除：': 'Edit/Delete:',
    '在列表中对已有私有网卡进行编辑或删除操作': 'In the list, you can edit or delete existing private NICs',
    'CIDR填写说明': 'CIDR Format Guildelines',
    '格式为': 'The format is',
    '例如：': 'For example:',
    '→ 可分配 IP 范围：': '→ Allocatable IP ranges:',
    '请避免与设备所在局域网网段冲突，建议使用': 'Please avoid IPs overlapping with the device LAN. Recommended blocks:',
    '范围内的地址': 'range',
    '公有网卡（MacVlan）说明': 'Public NIC (MacVlan) Guide',
    '公有网卡（MacVlan）模式下，虚拟机/容器将直接使用设备所在局域网的网关和子网，与物理设备处于同一网段。': 'In Public NIC (MacVlan) mode, VMs utilize the gateway and subnet of the host device explicitly, positioning themselves on the EXACT local network.',
    '优点：': 'Advantages:',
    '云机拥有局域网内真实IP，访问局域网资源更方便': 'VM acquires a real IP on the LAN, facilitating easy local access',
    '网络延迟更低，接近物理直连': 'Provides ultra-low network latencies akin to bare-metal linkage',
    '限制：': 'Restrictions:',
    '云机之间': 'Between VMs',
    '无法进行网络隔离': 'Network Isolation is IMPOSSIBLE',
    '设置后': 'Once configured,',
    '无法使用节点管理的IP代理功能': 'Node Management IP Proxies become completely UNAVAILABLE',
    '与"节点管理""节点分配"功能互斥': 'Mutually exclusive against both "Node Management" and "Node Allocation"',
    '查看网卡：': 'View NIC:',
    '右侧面板展示该设备物理网卡的当前配置（网关、子网掩码等）': 'The right panel illustrates current hardware network metadata (Gateways, Subnets)',
    '设置MacVlan IP：': 'Set MacVlan IP:',
    '为虚拟机/容器分配一个与设备同网段的IP地址，点击"设置MacVlan IP"后保存': 'Assign a sibling IP matching the host IP block. Afterwards, hit "Set MacVlan IP"',
    '更新MacVlan配置：': 'Update MacVlan Config:',
    '当设备更换局域网网段后，点击"更新MacVlan配置"将最新的网关/子网信息同步过来': 'When the physical machine migrates subnets, hit "Update MacVlan Config" to dynamically bridge the subnets',
    '此操作会': 'This operation will',
    '自动关闭所有正在运行的虚拟机和容器': 'AUTOMATICALLY SHUTDOWN ANY RUNNING VIRUTAL MACHINES OR CONTAINERS',
    '，执行后需要手动重新为每个容器设置MacVlan IP': '. Subsequently, it requires manual reallocation of IPs on each guest context',
    '功能互斥警告': 'Mutual Exclusion Warning',
    '启用 MacVlan（公有网卡）后，': 'Upon engaging MacVlan (Public NIC),',
    '节点管理、节点分配功能将被禁用': 'Node allocation capabilities shall irrevocably be disabled',
    '。如需切换，请先在公有网卡中移除所有 MacVlan 配置后再使用节点管理功能。': '. To switch modes, unconditionally flush any MacVlan setup prior to using proxies.',
    'IP分配建议': 'IP Allocation Suggestions',
    '为每台云机分配的IP必须与设备在同一网段，且不能与其他设备冲突': 'Static IPs distributed amongst the guests must exist within the gateway\'s broadcast boundaries, circumventing conflicts.',
    '建议从网段末尾开始分配，避免与 DHCP 自动分配的IP冲突': 'Allocate tail-end IPs consecutively to preclude standard DHCP assignment pools',
    '更换局域网网段后务必及时执行"更新MacVlan配置"': 'Mandatory syncing of MacVlan configs is essential following router network topology substitutions'
};

// Sort by length descending to replace larger chunks first
const keys = Object.keys(dict).sort((a,b) => b.length - a.length);

for (const key of keys) {
    // Replace all occurrences in enBlock
    // Using string replacement or global regex if safe
    enBlock = enBlock.split(key).join(dict[key]);
}

// Convert the old zhBlock into v-if
const zhReplacement = zhBlock.replace('<div style="margin: 20px;', '<div v-if="$i18n.locale === \'zh-CN\'" style="margin: 20px;');

// Convert the enBlock to v-else
const enReplacement = enBlock.replace('<div style="margin: 20px;', '<div v-else style="margin: 20px;');

// Splice them back into the Vue file
const finalContent = text.substring(0, divStart) + zhReplacement + '\n' + enReplacement + text.substring(divEnd);

fs.writeFileSync(tp, finalContent);
console.log('SUCCESS: Injected English translation for Network guide.');
