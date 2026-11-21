export default interface IAccessTokenData {
    id: number,
    companyId: number,
    name: string,
    token: string,
    expiresAt: Date,
    createdAt: Date,
    updatedAt: Date,
    deletedAt: Date | null,
}

// the token as the master sends it
export interface AccessTokenJson {
    id: number,
    company_id: number,
    name: string,
    token: string,
    expires_at: string,
    created_at: string,
    updated_at: string,
    deleted_at?: string | null,
}

export const parseAccessTokenData = (json: AccessTokenJson): IAccessTokenData => {
    return {
        id: json.id,
        companyId: json.company_id,
        name: json.name,
        token: json.token,
        expiresAt: new Date(json.expires_at),
        createdAt: new Date(json.created_at),
        updatedAt: new Date(json.updated_at),
        deletedAt: json.deleted_at ? new Date(json.deleted_at) : null,
    }
}