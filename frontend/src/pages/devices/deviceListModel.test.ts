import {describe, expect, it} from 'vitest';
import {filterByPlatform, groupDevices, platformCounts, platformsOf} from './deviceListModel';
import {PlatformType} from '../../types/platform.type.enum';
import {device, node} from './deviceFixtures';

const phone = device({id: 1, platformType: PlatformType.Android});
const iphone = device({id: 2, platformType: PlatformType.iOS});
const iphone2 = device({id: 3, platformType: PlatformType.iOS});

describe('platforms', () => {
    it('counts the devices per platform and lists the platforms in enum order', () => {
        const counts = platformCounts([phone, iphone, iphone2]);
        expect(counts).toEqual({[PlatformType.Android]: 1, [PlatformType.iOS]: 2});
        expect(platformsOf(counts)).toEqual([PlatformType.iOS, PlatformType.Android]);
        expect(platformsOf({})).toEqual([]);
    });

    it('filters by one platform or keeps all', () => {
        expect(filterByPlatform([phone, iphone], 'all')).toHaveLength(2);
        expect(filterByPlatform([phone, iphone], PlatformType.iOS)).toEqual([iphone]);
    });
});

describe('groupDevices', () => {
    const connected = node({id: 5, name: 'zeta', status: 1});
    const offline = node({id: 6, name: 'Alpha', status: 0});
    const onZeta = device({id: 10, nodeId: 5, node: connected});
    const onAlpha = device({id: 11, nodeId: 6, node: offline});
    const onMaster = device({id: 12});

    it('puts connected nodes first, then the rest by name, the master among them', () => {
        const groups = groupDevices([onMaster, onAlpha, onZeta], [connected, offline]);
        expect(groups.map(g => g.title)).toEqual(['Node: zeta (Connected)', 'Node: Alpha (Disconnected)', 'Node: Master']);
    });

    it('groups the devices of one node together', () => {
        const other = device({id: 13, nodeId: 5, node: connected});
        expect(groupDevices([onZeta, other], [connected])[0].devices).toEqual([onZeta, other]);
    });

    it('treats a node the hub state does not know as disconnected', () => {
        expect(groupDevices([onZeta], null)[0].title).toBe('Node: zeta (Disconnected)');
    });

    it('has no group without devices', () => {
        expect(groupDevices([], [])).toEqual([]);
    });
});
