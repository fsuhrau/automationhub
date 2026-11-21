import {render, screen} from '@testing-library/react';
import React from 'react';
import {describe, expect, it, vi} from 'vitest';
import {AppProviders} from './providers';
import {fakeApi} from '../testing/fakeApi';
import {useApiClient} from '../api/ApiClientProvider';
import {useError} from '../ErrorProvider';
import {useHubState} from '../hooks/HubStateProvider';

const http = vi.hoisted(() => ({get: vi.fn()}));
vi.mock('../http-common', () => ({default: http}));

const Probe: React.FC = () => {
    const api = useApiClient();
    const {error} = useError();
    const {state} = useHubState();
    const [answer, setAnswer] = React.useState('-');
    React.useEffect(() => {
        api.get<string>('/x').then(setAnswer);
    }, [api]);
    return <p>{`answer=${answer} error=${error} projects=${state.projects}`}</p>;
};

describe('AppProviders', () => {
    it('gives the pages the client it is given', async () => {
        render(<AppProviders client={fakeApi({'GET /x': 'from fake'})}><Probe/></AppProviders>);
        expect(await screen.findByText('answer=from fake error=null projects=null')).toBeInTheDocument();
    });

    it('gives the pages the real client by default', async () => {
        http.get.mockResolvedValue({data: 'from server'});
        render(<AppProviders><Probe/></AppProviders>);
        expect(await screen.findByText('answer=from server error=null projects=null')).toBeInTheDocument();
        expect(http.get).toHaveBeenCalledWith('/x', undefined);
    });
});
