// The sections every comparison has: the app, its size, how it is signed and
// what the master warned about.
import {IAppBinaryData} from '../../types/app';
import {BundleInfo, Engine} from '../../types/bundle.info';
import {CompareSection} from '../../types/bundle.compare';
import {commonName} from './bundleDetailModel';
import {present, sides} from './compareRows';

const engine = (e: Engine | undefined): string | undefined => e && [e.name, e.version].filter(Boolean).join(' ');

const tags = (b: IAppBinaryData): string[] => b.tags ? b.tags.split(',') : [];

export const appSection = (a: IAppBinaryData, b: IAppBinaryData): CompareSection => {
    const s = sides<IAppBinaryData>(a, b);
    return {
        title: 'App', rows: [
            s.text('Name', x => x.info?.displayName || x.name),
            s.text('Identifier', x => x.info?.identifier || x.identifier),
            s.text('Version', x => x.info?.version || x.version),
            s.text('Build', x => x.info?.build),
            s.text('Format', x => x.info?.format),
            s.text('Platform', x => x.platform),
            s.text('Engine', x => engine(x.info?.engine)),
            s.text('Scripting backend', x => x.info?.engine?.backend),
            s.list('Tags', tags),
        ],
    };
};

export const sizeSection = (a: IAppBinaryData, b: IAppBinaryData): CompareSection => {
    const s = sides<IAppBinaryData>(a, b);
    return {
        title: 'Size', rows: [
            s.number('File size', x => x.size || undefined, 'bytes'),
            s.number('Installed size', x => x.info?.contents?.size, 'bytes'),
            s.number('Compressed', x => x.info?.contents?.compressed, 'bytes'),
            s.number('Files', x => x.info?.contents?.files, 'count'),
        ],
    };
};

const certificates = (i: BundleInfo): string[] | undefined => i.signing?.certificates?.map(c =>
    `${commonName(c.subject)} (${c.sha1.slice(0, 8)}…)${c.debug ? ', debug key' : ''}`);

const firstExpiry = (i: BundleInfo): string | undefined => i.signing?.certificates
    ?.map(c => new Date(c.notAfter).toISOString().slice(0, 10)).filter(present).sort()[0];

export const signingSection = (a: BundleInfo | undefined, b: BundleInfo | undefined): CompareSection => {
    const s = sides(a, b);
    return {
        title: 'Signing', rows: [
            s.list('Signature schemes', i => i.signing?.schemes),
            s.list('Certificates', certificates),
            s.text('First certificate expiry', firstExpiry),
        ],
    };
};

export const warningsSection = (a: BundleInfo | undefined, b: BundleInfo | undefined): CompareSection =>
    ({title: 'Warnings', rows: [sides(a, b).list('Warnings', i => i.warnings)]});
