// TypeScript 接口定义，与后端 Go 数据模型一一对应

/** 添加自定义 DNS 服务器的请求 */
export interface AddServerRequest {
  protocol: string; // "udp" | "dot" | "doh"
  address: string;
  tlsServerName: string;
  bootstrapIP: string;
}

/** 服务器列表项（前端展示用） */
export interface ServerInfo {
  name: string;
  address: string;
  protocol: string; // "udp" | "dot" | "doh"
  tlsServerName: string;
  isPreset: boolean;
  canApplyToSystem?: boolean; // 是否可直接写入系统 DNS
}

/** 域名列表项（前端展示用） */
export interface DomainInfo {
  domain: string;
  isPreset: boolean;
}

/** 测试参数 */
export interface TestParams {
  queries: number;
  warmup: number;
  concurrency: number;
  timeout: number;
}

/** 单个服务器的测试结果 */
export interface TestResultItem {
  name: string;
  address: string;
  protocol: string;
  medianLatencyMs: number;
  p95LatencyMs: number;
  successRate: number;
  rawSuccesses: number;
  successes: number;
  total: number;
  answerMismatches: number;
  score: number;
  isTimeout: boolean;
  canApplyToSystem?: boolean; // 是否可直接写入系统 DNS
}

/** 完整测试结果 */
export interface TestResultsData {
  items: TestResultItem[];
  testTime: string;
  preset: string;
  bestDNS: string;
}

/** 网络适配器信息 */
export interface NetworkAdapterInfo {
  name: string;
  interfaceIdx: number;
  status: string;
  ipAddresses: string[]; // 适配器自身的 IPv4 地址
  currentDNS: string[];
}

/** 测试进度 */
export interface BenchmarkProgress {
  completed: number;
  total: number;
  percent: number;
}
