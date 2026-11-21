import {useEffect, useState} from 'react';
import {IAppBinaryData, IAppData} from '../../../types/app';
import {AppAction, DeviceControlApi} from '../../../types/device.control.api';
import {PlatformType} from '../../../types/platform.type.enum';
import {usableBundles} from '../../../services/app.service';
import {useLoader} from '../../../hooks/useLoader';
import {defaultApp, runsOn} from '../deviceControlModel';
import {ErrorHandler, Runner} from './useGuardedAction';

export interface AppPickerInput {
    service: DeviceControlApi,
    apps: IAppData[],
    platform: PlatformType | undefined,
    // the app picked in the menu, offered first when it runs on the device
    menuAppId: number | null,
    lease: string | null,
}

export interface AppPicker {
    appId: number | '',
    chooseApp: (id: number | '') => void,
    fittingApps: IAppData[],
    otherApps: IAppData[],
    bundles: IAppBinaryData[],
    bundleId: number | '',
    chooseBundle: (id: number | '') => void,
    clearData: boolean,
    setClearData: (clear: boolean) => void,
    // install, start or stop the chosen bundle; clearing data only applies to stop
    act: (action: AppAction) => void,
}

// useAppPicker keeps which app and bundle are chosen for install, start and
// stop: the app of the menu (or the first that runs on the device) and its
// newest bundle until the user picks others.
export const useAppPicker = (input: AppPickerInput, run: Runner, report: ErrorHandler): AppPicker => {
    const {service, apps, platform, menuAppId, lease} = input;
    const [chosenApp, setChosenApp] = useState<number | '' | null>(null);
    const [chosenBundle, setChosenBundle] = useState<number | '' | null>(null);
    const [clearData, setClearData] = useState(false);
    const appId = chosenApp ?? defaultApp(apps, platform, menuAppId)?.id ?? '';

    // the bundles come with their app, so those of the previous app never show
    const loaded = useLoader(() => appId === '' ? null : service.bundles(appId).then(list => ({appId, list: usableBundles(list)})), [service, appId]);
    const bundles = loaded.data?.appId === appId ? loaded.data.list : [];
    const bundleId = chosenBundle ?? bundles[0]?.id ?? '';
    useEffect(() => {
        if (loaded.error) report(loaded.error);
    }, [loaded.error, report]);

    const fittingApps = platform === undefined ? apps : apps.filter(a => runsOn(a.platform, platform));
    return {
        appId, fittingApps, otherApps: apps.filter(a => !fittingApps.includes(a)), bundles, bundleId, clearData, setClearData,
        chooseApp: id => {
            setChosenApp(id);
            setChosenBundle(null);
        },
        chooseBundle: setChosenBundle,
        act: action => {
            if (!lease || bundleId === '') return;
            run(() => service.appAction(lease, action, {binaryId: bundleId, clearData: action === 'stop' && clearData}));
        },
    };
};
