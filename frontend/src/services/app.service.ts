import http from "../http-common";
import IAppData from "../types/app"

class AppDataService {
    getAll() {
        return http.get("/apps");
    }

    get(id: string) {
        return http.get(`/app/${id}`);
    }

    create(data: IAppData) {
        return http.post("/app", data);
    }

    update(data: IAppData, id: any) {
        return http.put(`/app/${id}`, data);
    }

    delete(id: any) {
        return http.delete(`/app/${id}`);
    }

    findByName(name: string) {
        return http.get(`/apps?name=${name}`);
    }
}

export default new AppDataService();