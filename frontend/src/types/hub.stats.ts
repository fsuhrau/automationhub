import ITestProtocolData from './test.protocol';

export default interface IHubStatsData {
    AppsCount: number,
    AppsStorageSize: number,
    DatabaseSize: number,
    SystemMemoryUsage: number,
    SystemUptime: number,
    TestsLastProtocols: ITestProtocolData[],
    TestsLastFailed: ITestProtocolData[],
    DeviceCount: number,
    DeviceBooted: number,
}
