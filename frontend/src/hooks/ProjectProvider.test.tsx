import {render, screen} from '@testing-library/react';
import React from 'react';
import {Route, Routes} from 'react-router-dom';
import {beforeEach, describe, expect, it} from 'vitest';
import {ProjectProvider, useProjectContext} from './ProjectProvider';
import {aProject} from '../testing/fixtures';
import {renderWithProviders} from '../testing/renderWithProviders';
import IProject from '../project/project';

const Probe: React.FC = () => {
    const {project, projectIdentifier} = useProjectContext();
    return <p>{`project=${project.name} id=${projectIdentifier}`}</p>;
};

const mount = (path: string, projects?: IProject[]): void => {
    renderWithProviders(<Routes>
        <Route path="/" element={<p>project list</p>}/>
        <Route path="/project/:project_identifier" element={<ProjectProvider><Probe/></ProjectProvider>}/>
    </Routes>, {url: path, hub: {projects}});
};

beforeEach(() => localStorage.clear());

describe('ProjectProvider', () => {
    it('provides the project of the URL and remembers it', async () => {
        mount('/project/two', [aProject('one', 'One'), aProject('two', 'Two')]);
        expect(await screen.findByText('project=Two id=two')).toBeInTheDocument();
        expect(localStorage.length).toBeGreaterThan(0);
    });

    it('waits while the projects load', () => {
        mount('/project/two');
        expect(screen.getByRole('progressbar', {hidden: true})).toBeInTheDocument();
    });

    it.each([['an unknown project', '/project/zzz'], ['an identifier built from nothing', '/project/undefined'], ['a null identifier', '/project/null']])(
        'sends %s to the project list', async (_name, path) => {
            mount(path, [aProject('one')]);
            expect(await screen.findByText('project list')).toBeInTheDocument();
        });
});

describe('useProjectContext without a provider', () => {
    it('has no project', () => {
        render(<Probe/>);
        expect(screen.getByText('project=undefined id=')).toBeInTheDocument();
    });
});
