import http from "../http-common";
import IDeviceData from "../types/device"

class DeviceDataService {
    getAll() {
        return http.get("/devices");
    }

    get(id: string) {
        return http.get(`/device/${id}`);
    }

    create(data: IDeviceData) {
        return http.post("/device", data);
    }

    update(data: IDeviceData, id: any) {
        return http.put(`/device/${id}`, data);
    }

    delete(id: any) {
        return http.delete(`/device/${id}`);
    }

    findByName(name: string) {
        return http.get(`/devices?name=${name}`);
    }
}

export default new DeviceDataService();