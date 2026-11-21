import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {describe, expect, it, vi} from 'vitest';
import DateTime, {formatDateTime} from './DateTime';
import HubLogo from './HubLogo';
import MenuButton from './MenuButton';
import PlatformTypeIcon from './PlatformTypeIcon';
import {TitleCard} from './title.card.component';
import TopBar from './TopBar';
import {PlatformType} from '../types/platform.type.enum';
import {renderInApp} from '../app/testing/renderInApp';
import {fakeAuthClient, signedOut} from '../app/testing/fakeAuth';

vi.mock('../services/test.service', () => ({getTest: vi.fn()}));

describe('formatDateTime', () => {
    it.each([
        ['null', null],
        ['undefined', undefined],
        ['an empty string', ''],
        ['text that is no date', 'not a date'],
    ])('is empty for %s', (_name, value) => {
        expect(formatDateTime(value)).toBe('');
    });

    it('formats with the default or a given format', () => {
        const date = new Date(2024, 0, 5, 7, 8, 9);
        expect(formatDateTime(date)).toBe('2024/01/05 07:08:09');
        expect(formatDateTime(date, 'DD.MM.YYYY')).toBe('05.01.2024');
    });

    it('renders as a component', () => {
        render(<p><DateTime value={new Date(2024, 0, 5, 7, 8, 9)} format="YYYY"/></p>);
        expect(screen.getByText('2024')).toBeInTheDocument();
    });
});

describe('HubLogo', () => {
    it.each([
        ['without style', undefined],
        ['with one style', {fontSize: 28}],
        ['with several styles', [{fontSize: 28}, {mr: 1}]],
    ])('renders %s, hidden from screen readers', (_name, sx) => {
        const {container} = render(<HubLogo sx={sx}/>);
        expect(container.querySelector('svg')).toHaveAttribute('aria-hidden', 'true');
    });
});

describe('MenuButton', () => {
    it('is a button that can show a badge', async () => {
        const click = vi.fn();
        const {container} = render(<MenuButton aria-label="Open" onClick={click} showBadge={true}/>);
        await userEvent.click(screen.getByRole('button', {name: 'Open'}));
        expect(click).toHaveBeenCalled();
        expect(container.querySelector('.MuiBadge-invisible')).toBeNull();
    });

    it('hides the badge by default', () => {
        const {container} = render(<MenuButton aria-label="Open"/>);
        expect(container.querySelector('.MuiBadge-invisible')).not.toBeNull();
    });
});

describe('PlatformTypeIcon', () => {
    it.each(Object.values(PlatformType).filter(v => typeof v === 'number'))('draws an icon for platform %s', platform => {
        const {container} = render(<PlatformTypeIcon platformType={platform as PlatformType}/>);
        expect(container.querySelector('svg')).not.toBeNull();
    });

    it('draws nothing for an unknown platform', () => {
        const {container} = render(<PlatformTypeIcon platformType={99 as PlatformType}/>);
        expect(container).toBeEmptyDOMElement();
    });
});

describe('TitleCard', () => {
    it('shows title, subtitle, actions and content', () => {
        render(<TitleCard title="Devices" titleElement={<i>extra</i>} subtitle="all of them" actions={<button>Add</button>}>content</TitleCard>);
        expect(screen.getByRole('heading', {name: 'Devices'})).toBeInTheDocument();
        expect(screen.getByText('extra')).toBeInTheDocument();
        expect(screen.getByText('all of them')).toBeInTheDocument();
        expect(screen.getByRole('button', {name: 'Add'})).toBeInTheDocument();
        expect(screen.getByText('content')).toBeInTheDocument();
    });

    it('needs nothing but content', () => {
        render(<TitleCard>only content</TitleCard>);
        expect(screen.queryByRole('heading')).toBeNull();
        expect(screen.getByText('only content')).toBeInTheDocument();
    });
});

describe('TopBar', () => {
    it('leads home, shows breadcrumbs and the signed in user', async () => {
        renderInApp(<TopBar/>, {path: '/profile'});
        expect(await screen.findByText('Ada Lovelace')).toBeInTheDocument();
        expect(screen.getByRole('link', {name: /Automation Hub/})).toHaveAttribute('href', '/');
        expect(screen.getByRole('navigation', {name: 'breadcrumb'})).toHaveTextContent('Profile');
    });

    it('shows no user when nobody is signed in', async () => {
        renderInApp(<TopBar/>, {path: '/', auth: fakeAuthClient(signedOut)});
        expect(await screen.findByRole('navigation', {name: 'breadcrumb'})).toBeInTheDocument();
        expect(screen.queryByText('Ada Lovelace')).toBeNull();
    });
});
