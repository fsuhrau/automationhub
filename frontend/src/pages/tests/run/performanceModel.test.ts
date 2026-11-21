import {describe, expect, it} from 'vitest';
import {performance, protocol} from './testkit/fixtures';
import {averageOf, bundleLabel, checkpointRows, compare, measured, METRICS} from './performanceModel';

describe('checkpointRows', () => {
    it('times each checkpoint since the previous one and skips the schedule', () => {
        const rows = checkpointRows([
            performance({checkpoint: 'schedule', runtime: 0}), performance({checkpoint: 'menu', runtime: 2}), performance({checkpoint: 'game', runtime: 5}),
        ], null);
        expect(rows).toEqual([
            {checkpoint: 'menu', at: 2, took: 2, earlierTook: null},
            {checkpoint: 'game', at: 5, took: 3, earlierTook: null},
        ]);
    });

    it('sets the first occurrence of the same checkpoint of an earlier run next to it', () => {
        const rows = checkpointRows([performance({runtime: 4})], [performance({runtime: 1}), performance({runtime: 9})]);
        expect(rows[0].earlierTook).toBe(1);
    });

    it('is empty without data', () => {
        expect(checkpointRows(undefined, undefined)).toEqual([]);
    });
});

describe('compare', () => {
    it.each([
        {value: null, usual: 10, higher: true, verdict: null},
        {value: undefined, usual: 10, higher: true, verdict: null},
        {value: Infinity, usual: 10, higher: true, verdict: null},
        {value: 10, usual: 0, higher: true, verdict: null},
        {value: 10, usual: null, higher: true, verdict: null},
        {value: 10, usual: Infinity, higher: true, verdict: null},
        {value: 10.2, usual: 10, higher: true, verdict: 'same'},
        {value: 12, usual: 10, higher: true, verdict: 'better'},
        {value: 12, usual: 10, higher: false, verdict: 'worse'},
        {value: 8, usual: 10, higher: true, verdict: 'worse'},
        {value: 8, usual: 10, higher: false, verdict: 'better'},
    ])('compares $value with $usual (higher is better $higher) as $verdict', ({value, usual, higher, verdict}) => {
        expect(compare(value, usual, higher)).toBe(verdict);
    });
});

describe('metrics', () => {
    const p = protocol({avgFps: 59.94, avgMem: 2048, avgCpu: 12.34, avgVertexCount: 1500, avgTriangles: 900,
        histAvgFps: 1, histAvgMem: 2, histAvgCpu: 3, histAvgVertexCount: 4, histAvgTriangles: 5});

    it('reads the average and the usual value of each metric', () => {
        expect(METRICS.map(m => [m.key, m.avg(p), m.usual(p)])).toEqual([
            ['fps', 59.94, 1], ['mem', 2048, 2], ['cpu', 12.34, 3], ['vertexCount', 1500, 4], ['triangles', 900, 5],
        ]);
    });

    it('formats every metric', () => {
        expect(METRICS.map(m => m.format(m.avg(p) as number))).toEqual(['59.9 fps', '2.00 GB', '12.3 %', '1.5k', '900']);
        expect(METRICS[1].format(512)).toBe('512 MB');
    });

    it('averageOf reads the average of the key', () => {
        expect(averageOf({avgFps: 1, avgMem: 2, avgCpu: 3, avgVertexCount: 4, avgTriangles: 5}, 'cpu')).toBe(3);
    });
});

describe('measured', () => {
    it('needs a sample above zero', () => {
        expect(measured([performance({cpu: 0}), performance({cpu: 3})], 'cpu')).toBe(true);
        expect(measured([performance({cpu: 0})], 'cpu')).toBe(false);
        expect(measured(null, 'cpu')).toBe(false);
    });
});

describe('bundleLabel', () => {
    it.each([
        [{name: 'Game', version: '1.2'}, 'Game 1.2'],
        [{name: 'Game', version: ''}, 'Game'],
        [null, ''],
        [undefined, ''],
    ])('labels %j as %j', (bundle, label) => {
        expect(bundleLabel(bundle)).toBe(label);
    });
});
