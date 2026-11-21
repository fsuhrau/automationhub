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
