import ITestProtocolData from './test.protocol';

export default interface IHubStatsData {
    appsCount: number,
    appsStorageSize: number,
    databaseSize: number,
    systemMemoryUsage: number,
    systemUptime: number,
    testsLastProtocols: ITestProtocolData[],
    testsLastFailed: ITestProtocolData[],
    deviceCount: number,
    deviceBooted: number,
}
