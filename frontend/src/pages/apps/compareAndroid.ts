import {AndroidInfo} from '../../types/bundle.info';
import {CompareSection} from '../../types/bundle.compare';
import {apiLevel, shortPermission} from './bundleDetailModel';
import {sides} from './compareRows';

const permission = (p: { name: string, maxSdk?: string }): string =>
    shortPermission(p.name) + (p.maxSdk ? ` (up to API ${p.maxSdk})` : '');

const feature = (f: { name: string, required: boolean }): string =>
    `${f.name.replace(/^android\.hardware\./, '')}${f.required ? '' : ' (optional)'}`;

const level = (l: string | undefined): string | undefined => l ? apiLevel(l) : undefined;

// androidSection compares what the manifests of two bundles say
export const androidSection = (a: AndroidInfo | undefined, b: AndroidInfo | undefined): CompareSection => {
    const s = sides(a, b);
    return {
        title: 'Android', rows: [
            s.text('Package', i => i.packageName),
            s.text('Version name', i => i.versionName),
            s.text('Version code', i => i.versionCode),
            s.text('Min SDK', i => level(i.minSdk)),
            s.text('Target SDK', i => level(i.targetSdk)),
            s.text('Max SDK', i => level(i.maxSdk)),
            s.text('Compile SDK', i => level(i.compileSdk)),
            s.text('Launch activity', i => i.launchActivity),
            s.text('OpenGL ES', i => i.gles),
            s.flag('Debuggable', i => i.debuggable),
            s.flag('Test only', i => i.testOnly),
            s.flag('Android TV', i => i.leanback),
            s.list('Permissions', i => i.permissions?.map(permission)),
            s.list('Hardware features', i => i.features?.map(feature)),
            s.list('ABIs', i => i.abis),
            s.list('Locales', i => i.locales),
            s.list('Screens', i => i.screens),
            s.list('Densities', i => i.densities),
        ],
    };
};
