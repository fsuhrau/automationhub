import IAccessTokenData from "../types/access.token";
import { IAppData } from "../types/app";

export default interface IProject {
    ID: number,
    Identifier: string,
    Name: string,
    CompanyID: number,
    AccessTokens: IAccessTokenData[],
    Apps: IAppData[],
}