import {describe, expect, it} from 'vitest';
import {DeviceConnectionType, getConnectionTypeName, getConnectionTypes} from './device.connection.type.enum';

describe('connection types', () => {
    it('lists the types with their names', () => {
        expect(getConnectionTypes()).toEqual([{id: '0', name: 'USB'}, {id: '1', name: 'Remote'}, {id: '2', name: 'HubNode'}]);
        expect(getConnectionTypeName(DeviceConnectionType.Remote)).toBe('Remote');
    });
});
