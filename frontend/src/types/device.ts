export default interface IDeviceData {
    id?: number | null,
    DeviceIdentifier: string,
    DeviceType: number,
    Name: string,
    RAM: number,
    SoC: string,
    DisplaySize: string,
    DPI: number,
    OSVersion: string,
    GPU: string,
    ABI: string,
    OpenGL_ES_Version: number,
    Status: number,
    Manager: string,
}