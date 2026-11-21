import {describe, expect, it} from 'vitest';
import IDeviceData from '../types/device';
import {IAppData, AppParameter} from '../types/app';
import IProject from '../project/project';
import {DeviceStateType} from '../types/deviceStateType';
import {HubState, HubStateAction, HubStateActions as A, hubReducer, InitialHubState} from './HubState';

const asDevice = (d: Partial<IDeviceData>): IDeviceData => d as IDeviceData;
const asApp = (a: Partial<IAppData>): IAppData => a as IAppData;
const asProject = (p: Partial<IProject>): IProject => p as IProject;
const param = (id: number, appId: number, name = 'p'): AppParameter & { appId: number } =>
    ({id, appId, name, type: {type: 'string', defaultValue: ''}});
const withState = (s: Partial<HubState>): HubState => ({...InitialHubState, ...s});

const user = {id: 1} as never;
const token = {id: 5, name: 't'} as never;
const node = {id: 3, identifier: 'n', name: 'node', status: 1};
const dev1 = asDevice({id: 1, status: DeviceStateType.Booted, isLocked: true, lockedBy: 'me'});
const dev2 = asDevice({id: 2});
const app1 = asApp({id: 1, name: 'one', parameter: [param(7, 1)]});
const app2 = asApp({id: 2, name: 'two'});
const proj = asProject({identifier: 'a', name: 'A'});

interface Row {
    name: string,
    before: HubState,
    action: HubStateAction,
    after: HubState,
}

const rows: Row[] = [
    {name: 'UsersUpdate', before: InitialHubState, action: {type: A.UsersUpdate, payload: [user]}, after: withState({users: [user]})},
    {name: 'AppsUpdate', before: InitialHubState, action: {type: A.AppsUpdate, payload: [app1]}, after: withState({apps: [app1]})},
    {name: 'AppAdd to none', before: InitialHubState, action: {type: A.AppAdd, payload: app1}, after: withState({apps: [app1]})},
    {name: 'AppAdd appends', before: withState({apps: [app1]}), action: {type: A.AppAdd, payload: app2}, after: withState({apps: [app1, app2]})},
    {
        name: 'AppAttributeUpdate', before: withState({apps: [app1, app2]}),
        action: {type: A.AppAttributeUpdate, payload: {appId: 2, attribute: 'name', value: 'renamed'}},
        after: withState({apps: [app1, {...app2, name: 'renamed'}]}),
    },
    {
        name: 'AppAttributeUpdate without apps', before: InitialHubState,
        action: {type: A.AppAttributeUpdate, payload: {appId: 2, attribute: 'name', value: 'x'}}, after: withState({apps: []}),
    },
    {
        name: 'AddAppParameter', before: withState({apps: [app1, app2]}),
        action: {type: A.AddAppParameter, payload: param(8, 2)},
        after: withState({apps: [app1, {...app2, parameter: [param(8, 2)]}]}),
    },
    {
        name: 'UpdateAppParameter', before: withState({apps: [app1]}),
        action: {type: A.UpdateAppParameter, payload: param(7, 1, 'changed')},
        after: withState({apps: [{...app1, parameter: [param(7, 1, 'changed')]}]}),
    },
    {
        name: 'UpdateAppParameter on an app without list', before: withState({apps: [app2]}),
        action: {type: A.UpdateAppParameter, payload: param(7, 2)}, after: withState({apps: [{...app2, parameter: []}]}),
    },
    {
        name: 'DeleteAppParameter', before: withState({apps: [app1]}),
        action: {type: A.DeleteAppParameter, payload: {appId: 1, id: 7}}, after: withState({apps: [{...app1, parameter: []}]}),
    },
    {
        name: 'DeleteAppParameter on an app without list', before: withState({apps: [app2]}),
        action: {type: A.DeleteAppParameter, payload: {appId: 2, id: 7}}, after: withState({apps: [{...app2, parameter: []}]}),
    },
    {
        name: 'ProjectAttributeUpdate', before: withState({projects: [proj, asProject({identifier: 'b'})]}),
        action: {type: A.ProjectAttributeUpdate, payload: {projectIdentifier: 'a', attribute: 'name', value: 'new'}},
        after: withState({projects: [{...proj, name: 'new'}, asProject({identifier: 'b'})]}),
    },
    {
        name: 'ProjectAttributeUpdate without projects', before: InitialHubState,
        action: {type: A.ProjectAttributeUpdate, payload: {projectIdentifier: 'a', attribute: 'name', value: 'new'}},
        after: withState({projects: []}),
    },
    {name: 'ProjectsUpdate', before: InitialHubState, action: {type: A.ProjectsUpdate, payload: [proj]}, after: withState({projects: [proj]})},
    {name: 'ToggleMobileOpen on', before: InitialHubState, action: {type: A.ToggleMobileOpen}, after: withState({mobileOpen: true})},
    {name: 'ToggleMobileOpen off', before: withState({mobileOpen: true}), action: {type: A.ToggleMobileOpen}, after: InitialHubState},
    {name: 'AccessTokensUpdate', before: InitialHubState, action: {type: A.AccessTokensUpdate, payload: [token]}, after: withState({accessTokens: [token]})},
    {name: 'AccessTokenAdd to none', before: InitialHubState, action: {type: A.AccessTokenAdd, payload: token}, after: withState({accessTokens: [token]})},
    {
        name: 'AccessTokenAdd appends', before: withState({accessTokens: [token]}),
        action: {type: A.AccessTokenAdd, payload: {id: 6} as never}, after: withState({accessTokens: [token, {id: 6} as never]}),
    },
    {name: 'AccessTokenDelete', before: withState({accessTokens: [token, {id: 6} as never]}), action: {type: A.AccessTokenDelete, payload: 5}, after: withState({accessTokens: [{id: 6} as never]})},
    {name: 'AccessTokenDelete without tokens', before: InitialHubState, action: {type: A.AccessTokenDelete, payload: 5}, after: withState({accessTokens: []})},
    {name: 'NodesUpdate', before: InitialHubState, action: {type: A.NodesUpdate, payload: [node]}, after: withState({nodes: [node]})},
    {name: 'NodeAdd to none', before: InitialHubState, action: {type: A.NodeAdd, payload: node}, after: withState({nodes: [node]})},
    {name: 'NodeAdd appends', before: withState({nodes: [node]}), action: {type: A.NodeAdd, payload: {...node, id: 4}}, after: withState({nodes: [node, {...node, id: 4}]})},
    {name: 'NodeDelete', before: withState({nodes: [node, {...node, id: 4}]}), action: {type: A.NodeDelete, payload: 3}, after: withState({nodes: [{...node, id: 4}]})},
    {name: 'NodeDelete without nodes', before: InitialHubState, action: {type: A.NodeDelete, payload: 3}, after: withState({nodes: []})},
    {name: 'UpdateDevices', before: InitialHubState, action: {type: A.UpdateDevices, payload: [dev1]}, after: withState({devices: [dev1]})},
    {
        name: 'UpdateDevice', before: withState({devices: [dev1, dev2]}),
        action: {type: A.UpdateDevice, payload: {...dev2, alias: 'x'}}, after: withState({devices: [dev1, {...dev2, alias: 'x'}]}),
    },
    {name: 'UpdateDevice without devices', before: InitialHubState, action: {type: A.UpdateDevice, payload: dev2}, after: withState({devices: []})},
    {
        name: 'UpdateDeviceState', before: withState({devices: [dev1, dev2]}),
        action: {type: A.UpdateDeviceState, payload: {deviceId: 2, deviceState: DeviceStateType.Shutdown}},
        after: withState({devices: [dev1, {...dev2, status: DeviceStateType.Shutdown}]}),
    },
    {
        name: 'UpdateDeviceState without devices', before: InitialHubState,
        action: {type: A.UpdateDeviceState, payload: {deviceId: 2, deviceState: DeviceStateType.Shutdown}}, after: withState({devices: []}),
    },
    {
        name: 'UnlockDevice', before: withState({devices: [dev1, dev2]}),
        action: {type: A.UnlockDevice, payload: {deviceId: 1}},
        after: withState({devices: [{...dev1, isLocked: false, lockedBy: undefined}, dev2]}),
    },
];

describe('hubReducer', () => {
    it.each(rows)('$name', ({before, action, after}) => {
        expect(hubReducer(before, action)).toEqual(after);
    });

    it('covers every action type', () => {
        const covered = new Set(rows.map(r => r.action.type));
        expect([...covered].sort()).toEqual(Object.values(A).sort());
    });

    it('does not change the state it is given', () => {
        const before = withState({apps: [app1]});
        const snapshot = structuredClone(before);
        hubReducer(before, {type: A.AppAttributeUpdate, payload: {appId: 1, attribute: 'name', value: 'z'}});
        expect(before).toEqual(snapshot);
    });

    it('leaves the state unchanged for an unknown action', () => {
        const state = withState({mobileOpen: true});
        expect(hubReducer(state, {type: 'Nope'} as unknown as HubStateAction)).toBe(state);
    });
});
