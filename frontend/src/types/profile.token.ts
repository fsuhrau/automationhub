export interface IProfile {
    id: number,
    name: string,
    login: string,
    email: string,
}

export interface IPersonalToken {
    id: number,
    name: string,
    // last characters of the token, for recognizing it
    hint: string,
    createdAt: string,
    expiresAt: string | null,
    lastUsedAt: string | null,
}

// the value is only returned once, when the token is created
export interface ICreatedPersonalToken extends IPersonalToken {
    token: string,
}
