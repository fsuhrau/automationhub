import {render, screen} from '@testing-library/react';
import React from 'react';
import {MemoryRouter, Route, Routes, useLocation} from 'react-router-dom';
import {beforeEach, describe, expect, it} from 'vitest';
import {HubStateProvider} from './HubStateProvider';
import {ProjectProvider} from './ProjectProvider';
import ApplicationPage, {useApplicationContext} from './ApplicationProvider';
import {aProject, anApp, SeedHub} from '../app/testing/seedHub';
import {rememberApp} from './useSelectedApp';
import {IAppData} from '../types/app';

const Page: React.FC = () => {
    const {appId} = useApplicationContext();
    const location = useLocation();
    return <p>{`appId=${appId} at=${location.pathname}${location.search}`}</p>;
};

const mount = (path: string, apps?: IAppData[]): void => {
    render(<HubStateProvider><SeedHub projects={[aProject('p')]} apps={apps}><MemoryRouter initialEntries={[path]}>
        <Routes>
            <Route path="/project/:project_identifier/settings" element={<Page/>}/>
            <Route path="/project/:project_identifier/:appId" element={<ProjectProvider><ApplicationPage/></ProjectProvider>}>
                <Route path="tests" element={<Page/>}/>
            </Route>
        </Routes>
    </MemoryRouter></SeedHub></HubStateProvider>);
};

beforeEach(() => localStorage.clear());

describe('ApplicationProvider', () => {
    it('provides the app of the URL', async () => {
        mount('/project/p/app:2/tests', [anApp(1), anApp(2)]);
        expect(await screen.findByText('appId=2 at=/project/p/app:2/tests')).toBeInTheDocument();
    });

    it('keeps an app of the URL that is not known, so the page can show its error', async () => {
        mount('/project/p/app:9/tests', [anApp(1)]);
        expect(await screen.findByText('appId=9 at=/project/p/app:9/tests')).toBeInTheDocument();
    });

    it('falls back to the remembered app when the URL has no app id', async () => {
        rememberApp(2);
        mount('/project/p/other/tests', [anApp(1), anApp(2)]);
        expect(await screen.findByText('appId=2 at=/project/p/other/tests')).toBeInTheDocument();
    });

    it('waits while the apps load', async () => {
        mount('/project/p/app:2/tests');
        expect(await screen.findByRole('progressbar', {hidden: true})).toBeInTheDocument();
    });

    it('sends a project without apps to the apps settings', async () => {
        mount('/project/p/app:1/tests', []);
        expect(await screen.findByText('appId=null at=/project/p/settings?tab=apps')).toBeInTheDocument();
    });
});

describe('useApplicationContext without a provider', () => {
    it('has no app', () => {
        render(<MemoryRouter><Page/></MemoryRouter>);
        expect(screen.getByText('appId=null at=/')).toBeInTheDocument();
    });
});
