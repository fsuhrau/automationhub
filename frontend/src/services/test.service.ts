import http from "../http-common";
import ITestData from "../types/test"

class TestDataService {
    getAll() {
        return http.get("/tests");
    }

    get(id: string) {
        return http.get(`/test/${id}`);
    }

    create(data: ITestData) {
        return http.post("/test", data);
    }

    update(data: ITestData, id: any) {
        return http.put(`/test/${id}`, data);
    }

    delete(id: any) {
        return http.delete(`/test/${id}`);
    }

    findByName(name: string) {
        return http.get(`/tests?name=${name}`);
    }

    executeTest(id: number | null | undefined, appid: number, devices: Array<number>) {
        return http.post(`/test/${id}/run`, {
            AppID: appid,
            Devices: devices,
        });
    }
}

export default new TestDataService();