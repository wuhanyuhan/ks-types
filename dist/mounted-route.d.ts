export const PMMethodMountedRouteChanged = "keystone.mounted.route.changed";
export const PMMethodMountedRouteRestore = "keystone.mounted.route.restore";

export interface MountedRouteChangedMessage {
  type: typeof PMMethodMountedRouteChanged;
  version: number;
  appId: string;
  path: string;
  hash?: string;
  title?: string;
  replace?: boolean;
}

export interface MountedRouteRestoreMessage {
  type: typeof PMMethodMountedRouteRestore;
  version: number;
  path: string;
  hash?: string;
  replace?: boolean;
}
