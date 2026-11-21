import IProject from '../../project/project';
import IAccessTokenData from '../../types/access.token';
import {AppParameter, IAppData} from '../../types/app';
import {INodeData} from '../../types/node';
import {PlatformType} from '../../types/platform.type.enum';

export const project = (patch: Partial<IProject> = {}): IProject =>
    ({id: 1, identifier: 'p1', name: 'Heroes', companyId: 1, accessTokens: [], apps: [], ...patch});

export const app = (patch: Partial<IAppData> = {}): IAppData =>
    ({id: 1, name: 'Game Android', identifier: 'com.game', platform: PlatformType.Android, parameter: [], ...patch}) as IAppData;

export const parameter = (patch: Partial<AppParameter> = {}): AppParameter =>
    ({id: 10, name: 'MarketId', type: {type: 'string', defaultValue: 'eu'}, ...patch});

export const accessToken = (patch: Partial<IAccessTokenData> = {}): IAccessTokenData =>
    ({id: 5, name: 'CI nightly', token: 'abcdef123456789', expiresAt: null, createdAt: new Date('2024-02-03'), ...patch}) as IAccessTokenData;

export const node = (patch: Partial<INodeData> = {}): INodeData =>
    ({id: 3, identifier: 'node-1', name: 'Mac mini QA', status: 1, ...patch});
