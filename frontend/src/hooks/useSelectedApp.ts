import {useParams} from "react-router-dom";
import {useHubState} from "./HubStateProvider";
import {IAppData} from "../types/app";

const STORAGE_KEY = 'appId';

export const rememberApp = (appId: number) => {
    try {
        localStorage.setItem(STORAGE_KEY, `${appId}`);
    } catch {
        // storage is a convenience only
    }
};

const rememberedApp = (): number | null => {
    try {
        const value = localStorage.getItem(STORAGE_KEY);
        return value ? +value.replace('app:', '') : null;
    } catch {
        return null;
    }
};

// parseAppParam turns the ":appId" route segment ("app:12") into the id.
export const parseAppParam = (param?: string): number | null => {
    if (!param || !param.startsWith('app:')) {
        return null;
    }
    const id = +param.substring('app:'.length);
    return isNaN(id) ? null : id;
};

export interface SelectedApp {
    // the app of the page, or the one picked last, or the first of the project
    app: IAppData | null,
    // true when the current page belongs to the app (its id is in the URL)
    inUrl: boolean,
}

// useSelectedApp is the one place that decides which app the app scoped
// navigation points to. The app in the URL always wins; outside app pages the
// last picked app of this project is used, so the menu never points to an app
// of another project.
export const useSelectedApp = (): SelectedApp => {
    const {state} = useHubState();
    const {appId: param} = useParams<{ appId: string }>();
    const apps = state.apps ?? [];

    const fromUrl = parseAppParam(param);
    const urlApp = fromUrl !== null ? apps.find(a => a.id === fromUrl) : undefined;
    if (urlApp) {
        return {app: urlApp, inUrl: true};
    }
    const stored = rememberedApp();
    const app = apps.find(a => a.id === stored) ?? apps[0] ?? null;
    return {app, inUrl: false};
};
