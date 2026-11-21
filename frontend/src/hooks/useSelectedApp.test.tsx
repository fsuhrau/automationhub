import {render, screen} from '@testing-library/react';
import React from 'react';
import {MemoryRouter, Route, Routes} from 'react-router-dom';
import {beforeEach, describe, expect, it, vi} from 'vitest';
import {HubStateProvider} from './HubStateProvider';
import {parseAppParam, rememberApp, useSelectedApp} from './useSelectedApp';
import {anApp, SeedHub} from '../app/testing/seedHub';

const Probe: React.FC = () => {
    const {app, inUrl} = useSelectedApp();
    return <p>{`app=${app?.name ?? '-'} inUrl=${inUrl}`}</p>;
};

const mount = async (path: string, apps = [anApp(1, 'One'), anApp(2, 'Two')]): Promise<void> => {
    render(<HubStateProvider><SeedHub apps={apps}><MemoryRouter initialEntries={[path]}>
        <Routes><Route path="/project/:projectId/:appId?/*" element={<Probe/>}/></Routes>
    </MemoryRouter></SeedHub></HubStateProvider>);
    await screen.findByText(/^app=/);
};

beforeEach(() => localStorage.clear());

describe('parseAppParam', () => {
    it.each([
        ['app:12', 12],
        ['12', null],
        [undefined, null],
        ['app:x', null],
    ])('turns %s into %s', (param, id) => {
        expect(parseAppParam(param)).toBe(id);
    });
});

describe('useSelectedApp', () => {
    it('takes the app of the URL', async () => {
        await mount('/project/p/app:2/tests');
        expect(screen.getByText('app=Two inUrl=true')).toBeInTheDocument();
    });

    it('takes the remembered app outside app pages', async () => {
        rememberApp(2);
        await mount('/project/p');
        expect(screen.getByText('app=Two inUrl=false')).toBeInTheDocument();
    });

    it('takes the first app when the remembered one is not in the project', async () => {
        rememberApp(9);
        await mount('/project/p/app:7/tests');
        expect(screen.getByText('app=One inUrl=false')).toBeInTheDocument();
    });

    it('has no app in a project without apps', async () => {
        await mount('/project/p', []);
        expect(screen.getByText('app=- inUrl=false')).toBeInTheDocument();
    });

    it('copes with a remembered app that is not a number', async () => {
        localStorage.setItem('appId', 'app:2');
        await mount('/project/p');
        expect(screen.getByText('app=Two inUrl=false')).toBeInTheDocument();
    });

    it('works when the browser storage is blocked', async () => {
        const blocked = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
            throw new Error('blocked');
        });
        const blockedSet = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new Error('blocked');
        });
        rememberApp(2);
        await mount('/project/p');
        expect(screen.getByText('app=One inUrl=false')).toBeInTheDocument();
        blocked.mockRestore();
        blockedSet.mockRestore();
    });
});
