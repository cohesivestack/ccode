import type * as V3_0 from "./v3_0";
import type * as V3_1 from "./v3_1";
import type * as V3_2 from "./v3_2";

/** Supported OpenAPI major/minor versions. */
export type Version = "3.0" | "3.1" | "3.2";

/** Independent ergonomic model for each specification version. */
export interface ApiModelByVersion {
  "3.0": V3_0.ApiModel;
  "3.1": V3_1.ApiModel;
  "3.2": V3_2.ApiModel;
}

/** Selects one version, or all supported versions when omitted. */
export type ApiModel<V extends Version = Version> = ApiModelByVersion[V];
