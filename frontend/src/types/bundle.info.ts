// What the master reads out of an uploaded apk, ipa or zipped app; mirrors
// app.BundleInfo on the master.

export interface Engine {
    name: string,
    version?: string,
    // IL2CPP or Mono for Unity
    backend?: string,
}

export interface AndroidPermission {
    name: string,
    maxSdk?: string,
}

export interface AndroidFeature {
    name: string,
    required: boolean,
}

export interface AndroidInfo {
    packageName: string,
    versionCode?: string,
    versionName?: string,
    minSdk?: string,
    targetSdk?: string,
    maxSdk?: string,
    compileSdk?: string,
    compileSdkCodename?: string,
    platformBuildVersion?: string,
    label?: string,
    launchActivity?: string,
    permissions?: AndroidPermission[],
    features?: AndroidFeature[],
    abis?: string[],
    locales?: string[],
    densities?: string[],
    screens?: string[],
    gles?: string,
    debuggable: boolean,
    testOnly: boolean,
    leanback: boolean,
}

export interface UsageDescription {
    key: string,
    text: string,
}

export interface ProvisioningProfile {
    name: string,
    uuid?: string,
    appIdName?: string,
    teamName?: string,
    teamId?: string,
    kind: 'development' | 'ad-hoc' | 'enterprise' | 'app-store',
    createdAt: string,
    expiresAt: string,
    allDevices: boolean,
    devices?: string[],
    entitlements?: Record<string, unknown>,
}

export interface AppleInfo {
    platform?: string,
    minOs?: string,
    sdk?: string,
    xcode?: string,
    xcodeBuild?: string,
    buildMachineOs?: string,
    bundleName?: string,
    executable?: string,
    deviceFamily?: string[],
    requiredCapabilities?: string[],
    orientations?: string[],
    backgroundModes?: string[],
    urlSchemes?: string[],
    usageDescriptions?: UsageDescription[],
    extensions?: string[],
    frameworks?: string[],
    fileSharing: boolean,
    profile?: ProvisioningProfile,
}

export interface Certificate {
    subject: string,
    issuer: string,
    serial: string,
    notBefore: string,
    notAfter: string,
    sha1: string,
    sha256: string,
    debug: boolean,
}

export interface Signing {
    schemes?: string[],
    certificates?: Certificate[],
}

export interface SizeEntry {
    path: string,
    files?: number,
    size: number,
    compressed: number,
}

export interface Contents {
    files: number,
    size: number,
    compressed: number,
    groups?: SizeEntry[],
    largest?: SizeEntry[],
}

export interface BundleInfo {
    format: string,
    displayName?: string,
    identifier?: string,
    version?: string,
    build?: string,
    icon?: string,
    engine?: Engine,
    android?: AndroidInfo,
    apple?: AppleInfo,
    signing?: Signing,
    contents?: Contents,
    warnings?: string[],
    readAt: string,
}
