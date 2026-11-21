import {Dayjs} from 'dayjs';
import {INodeData} from './node';
import {PlatformType} from './platform.type.enum';

export interface NewAccessTokenRequest {
    name: string,
    expiresAt: Dayjs | null,
}

export interface NewNodeRequest {
    name: string,
}

// a node with its token; the token only comes with creating a node and
// issuing a new token
export type NodeWithToken = INodeData & { token: string };

export interface NewAppRequest {
    name: string,
    platform: PlatformType,
    identifier: string,
}

// the body the master stores for an app parameter
export interface ParameterRequest {
    name: string,
    type: { type: 'string' | 'option', defaultValue: string, options?: string[] },
}
