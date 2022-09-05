import IAccessTokenData from "../types/access.token";
import IAppData from "../types/app";

export default interface IProject {
    ID: string,
    CompanyID: number,
    Name: string,
    AccessTokens: IAccessTokenData[],
    Apps: IAppData[],
}