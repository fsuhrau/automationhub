// Comparing two bundles without React: which attributes changed from the
// older bundle (a) to the newer one (b), which entries of lists like
// permissions or frameworks came and went, and where the size moved.
import {IAppBinaryData} from '../../types/app';
import {AppleInfo, BundleInfo} from '../../types/bundle.info';
import {CompareSection, Comparison} from '../../types/bundle.compare';
import {androidSection} from './compareAndroid';
import {appleSection, profileSection} from './compareApple';
import {appSection, signingSection, sizeSection, warningsSection} from './compareGeneral';
import {sizes} from './compareSizes';

const both = <T, >(a: T | undefined, b: T | undefined): boolean => a !== undefined || b !== undefined;

type MaybeInfo = BundleInfo | undefined;

const androidSections = (a: MaybeInfo, b: MaybeInfo): CompareSection[] =>
    both(a?.android, b?.android) ? [androidSection(a?.android, b?.android)] : [];

const profileSections = (a: AppleInfo | undefined, b: AppleInfo | undefined): CompareSection[] =>
    both(a?.profile, b?.profile) ? [profileSection(a?.profile, b?.profile)] : [];

const appleSections = (a: MaybeInfo, b: MaybeInfo): CompareSection[] =>
    both(a?.apple, b?.apple) ? [appleSection(a?.apple, b?.apple), ...profileSections(a?.apple, b?.apple)] : [];

const signingSections = (a: MaybeInfo, b: MaybeInfo): CompareSection[] =>
    both(a?.signing, b?.signing) ? [signingSection(a, b)] : [];

// platformSections are the sections that exist only when a bundle has them
const platformSections = (a: MaybeInfo, b: MaybeInfo): CompareSection[] =>
    [...androidSections(a, b), ...appleSections(a, b), ...signingSections(a, b)];

const countChanges = (sections: CompareSection[]): number =>
    sections.reduce((n, s) => n + s.rows.filter(r => r.change !== 'same').length, 0);

// compareBundles compares bundle a with bundle b; a is usually the older one
export const compareBundles = (a: IAppBinaryData, b: IAppBinaryData): Comparison => {
    const [ia, ib] = [a.info, b.info];
    const sections = [appSection(a, b), sizeSection(a, b), ...platformSections(ia, ib), warningsSection(ia, ib)];
    return {
        sections,
        parts: sizes(ia?.contents?.groups, ib?.contents?.groups),
        files: sizes(ia?.contents?.largest, ib?.contents?.largest, true),
        changes: countChanges(sections),
    };
};

// onlyChanges drops the rows and sections without a change
export const onlyChanges = (sections: CompareSection[]): CompareSection[] =>
    sections.map(s => ({...s, rows: s.rows.filter(r => r.change !== 'same')})).filter(s => s.rows.length > 0);

// hideEmpty drops rows that neither bundle has
export const hideEmpty = (sections: CompareSection[]): CompareSection[] =>
    sections.map(s => ({
        ...s, rows: s.rows.filter(r => r.kind === 'value' ? r.a !== undefined || r.b !== undefined : r.same + r.added.length + r.removed.length > 0),
    })).filter(s => s.rows.length > 0);

// previousOf is the bundle uploaded before the given one, for the default
// comparison; bundles are compared to the one before them
export const previousOf = (bundles: IAppBinaryData[], id: number): IAppBinaryData | undefined =>
    bundles.filter(b => b.id < id).sort((x, y) => y.id - x.id)[0];
