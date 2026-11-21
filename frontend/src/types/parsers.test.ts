import {describe, expect, it} from 'vitest';
import {parseAccessTokenData} from './access.token';
import {parseNodeData} from './node';
import {DeviceConnectionType, getConnectionTypeName, getConnectionTypes} from './device.connection.type.enum';

describe('parseAccessTokenData', () => {
    const json = {
        id: 1, company_id: 2, name: 't', token: 'secret',
        expires_at: '2030-01-01T00:00:00Z', created_at: '2020-01-01T00:00:00Z', updated_at: '2021-01-01T00:00:00Z',
    };

    it('maps the fields and dates of the master', () => {
        expect(parseAccessTokenData(json)).toEqual({
            id: 1, companyId: 2, name: 't', token: 'secret',
            expiresAt: new Date('2030-01-01T00:00:00Z'), createdAt: new Date('2020-01-01T00:00:00Z'),
            updatedAt: new Date('2021-01-01T00:00:00Z'), deletedAt: null,
        });
    });

    it('reads the deletion date when there is one', () => {
        expect(parseAccessTokenData({...json, deleted_at: '2022-01-01T00:00:00Z'}).deletedAt).toEqual(new Date('2022-01-01T00:00:00Z'));
    });
});

describe('parseNodeData', () => {
    it('keeps the node fields', () => {
        expect(parseNodeData({id: 1, identifier: 'n', name: 'node', status: 2, extra: true} as never))
            .toEqual({id: 1, identifier: 'n', name: 'node', status: 2});
    });
});

describe('connection types', () => {
    it('lists the types with their names', () => {
        expect(getConnectionTypes()).toEqual([{id: '0', name: 'USB'}, {id: '1', name: 'Remote'}, {id: '2', name: 'HubNode'}]);
        expect(getConnectionTypeName(DeviceConnectionType.Remote)).toBe('Remote');
    });
});
