// 行业库 / 政策库 / 全市场缓存 / Python 依赖 / 通用对话框 / 系统通知
import { Go_ } from './wrap'

export const {
  UpdatePolicyLibrary,
  ReloadPolicyLibrary,
  SaveDefaultPolicyLibrary,
  GetPolicyLibraryMeta,
  UpdateIndustryDatabase,
  ReloadIndustryDatabase,
  RefreshIndustryBaselines,
  GetIndustryDBMeta,
  GetIndustryMetrics,
  GetIndustryTaskStatus,
  InitIndustryDatabase,
  GetMarketCacheStatus,
  RefreshMarketCache,
  CheckPythonDependencies,
  InstallPythonDependencies,
  HasPythonDepsChecked,
  MarkPythonDepsChecked,
  ConfirmDialog,
  SendNotification,
} = Go_
