import React from 'react';
import {RenderResult} from '@testing-library/react';
import {Route, Routes} from 'react-router-dom';
import {ProjectProvider} from '../../hooks/ProjectProvider';
import IProject from '../../project/project';
import {ProviderOptions, renderWithProviders} from '../../testing/renderWithProviders';
import {project} from './settingsFixtures';

const DEFAULT_PROJECTS = [project()];

// renderInProject renders a settings page at the project route, the project p1 known to the hub
export const renderInProject = (page: React.ReactElement, options: ProviderOptions & { projects?: IProject[] } = {}): RenderResult => {
    const {projects = DEFAULT_PROJECTS, url = '/project/p1/settings', ...rest} = options;
    return renderWithProviders(
        <Routes><Route path="/project/:project_identifier/*" element={<ProjectProvider>{page}</ProjectProvider>}/></Routes>,
        {...rest, url, hub: {projects}},
    );
};
