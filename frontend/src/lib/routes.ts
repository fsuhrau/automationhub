// Every path of App.tsx in one place. Pages build links with these instead
// of template strings; breadcrumbs.ts parses the same shapes.

const enc = encodeURIComponent;

type Id = number | string;

export interface RunRoutes {
    show: () => string,
    replay: () => string,
    protocol: (protocolId: Id) => string,
}

export interface TestRoutes {
    show: () => string,
    edit: () => string,
    lastRun: () => string,
    run: (runId: Id) => RunRoutes,
}

export interface AppRoutes {
    home: () => string,
    tests: () => string,
    newTest: () => string,
    test: (testId: Id) => TestRoutes,
    bundles: () => string,
    compareBundles: () => string,
    bundle: (bundleId: Id) => string,
    results: () => string,
    performance: () => string,
}

export interface DeviceRoutes {
    show: () => string,
    control: () => string,
    edit: () => string,
}

export interface ProjectRoutes {
    overview: () => string,
    settings: (tab?: string) => string,
    users: () => string,
    devices: () => string,
    device: (deviceId: Id) => DeviceRoutes,
    app: (appId: Id) => AppRoutes,
}

const run = (testBase: string, runId: Id): RunRoutes => {
    const base = `${testBase}/run/${runId}`;
    return {
        show: (): string => base,
        replay: (): string => `${base}/replay`,
        protocol: (protocolId: Id): string => `${base}/${protocolId}`,
    };
};

const test = (appBase: string, testId: Id): TestRoutes => {
    const base = `${appBase}/test/${testId}`;
    return {
        show: (): string => base,
        edit: (): string => `${base}/edit`,
        lastRun: (): string => `${base}/runs/last`,
        run: (runId: Id) => run(base, runId),
    };
};

const app = (projectBase: string, appId: Id): AppRoutes => {
    const base = `${projectBase}/app:${appId}`;
    return {
        home: (): string => `${base}/home`,
        tests: (): string => `${base}/tests`,
        newTest: (): string => `${base}/test/new`,
        test: (testId: Id) => test(base, testId),
        bundles: (): string => `${base}/bundles`,
        compareBundles: (): string => `${base}/bundles/compare`,
        bundle: (bundleId: Id): string => `${base}/bundle/${bundleId}`,
        results: (): string => `${base}/results`,
        performance: (): string => `${base}/performance`,
    };
};

const device = (projectBase: string, deviceId: Id): DeviceRoutes => {
    const base = `${projectBase}/device/${deviceId}`;
    return {
        show: (): string => base,
        control: (): string => `${base}/control`,
        edit: (): string => `${base}/edit`,
    };
};

const project = (identifier: string): ProjectRoutes => {
    const base = `/project/${enc(identifier)}`;
    return {
        overview: (): string => base,
        settings: (tab?: string): string => tab ? `${base}/settings?tab=${enc(tab)}` : `${base}/settings`,
        users: (): string => `${base}/users`,
        devices: (): string => `${base}/devices`,
        device: (deviceId: Id) => device(base, deviceId),
        app: (appId: Id) => app(base, appId),
    };
};

export const routes = {
    login: (): string => '/login',
    signup: (): string => '/signup',
    projects: (): string => '/',
    profile: (): string => '/profile',
    project,
};
