import {fireEvent, render, screen} from '@testing-library/react';
import {describe, expect, it} from 'vitest';
import React from 'react';
import {HubStateActions} from '../application/HubState';
import {HubStateProvider, useHubState} from './HubStateProvider';

const Consumer: React.FC<{ name: string }> = ({name}) => {
    const {state} = useHubState();
    return <p>{name}:{String(state.mobileOpen)}</p>;
};

const Toggler: React.FC = () => {
    const {dispatch} = useHubState();
    return <button onClick={() => dispatch({type: HubStateActions.ToggleMobileOpen})}>toggle</button>;
};

const toggle = (): void => {
    fireEvent.click(screen.getByText('toggle'));
};

describe('HubStateProvider', () => {
    it('starts with the initial state', () => {
        render(<HubStateProvider><Consumer name="a"/></HubStateProvider>);
        expect(screen.getByText('a:false')).toBeInTheDocument();
    });

    it('updates every consumer when an action is dispatched', () => {
        render(<HubStateProvider><Toggler/><Consumer name="a"/><Consumer name="b"/></HubStateProvider>);
        toggle();
        expect(screen.getByText('a:true')).toBeInTheDocument();
        expect(screen.getByText('b:true')).toBeInTheDocument();
    });

    it('ignores dispatches without a provider', () => {
        render(<><Toggler/><Consumer name="a"/></>);
        toggle();
        expect(screen.getByText('a:false')).toBeInTheDocument();
    });
});
