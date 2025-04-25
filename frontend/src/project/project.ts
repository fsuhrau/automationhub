import IAccessTokenData from "../types/access.token";
import { IAppData } from "../types/app";

export default interface IProject {
    id: number,
    identifier: string,
    name: string,
    companyId: number,
    accessTokens: IAccessTokenData[],
    apps: IAppData[],
}