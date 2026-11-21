// Where a bundle lives: the project and the app, plus the bundle itself.

export interface AppRef {
    projectId: string,
    appId: number,
}

export interface BundleRef extends AppRef {
    id: number,
}
