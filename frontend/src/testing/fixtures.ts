import IProject from '../project/project';
import IDeviceData from '../types/device';
import {IAppData} from '../types/app';
import {PlatformType} from '../types/platform.type.enum';

export const aProject = (identifier = 'p1', name = identifier): IProject =>
    ({id: 1, identifier, name, companyId: 1, accessTokens: [], apps: []});

export const anApp = (id: number, name = `app ${id}`): IAppData =>
    ({id, name, identifier: `app.${id}`, projectId: 1, platform: PlatformType.Android} as IAppData);

export const aDevice = (fields: Partial<IDeviceData> = {}): IDeviceData => ({id: 1, name: 'phone', alias: '', ...fields} as IDeviceData);
