import {expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {Chips} from './Chips';

it('shows nothing without items', () => {
    const {container} = render(<Chips items={[]}/>);
    expect(container).toBeEmptyDOMElement();
});

it('shows the first items and the rest on click', async () => {
    render(<Chips items={['a', 'b', 'c']} max={2}/>);
    expect(screen.queryByText('c')).not.toBeInTheDocument();
    await userEvent.click(screen.getByText('+1 more'));
    expect(screen.getByText('c')).toBeInTheDocument();
    expect(screen.queryByText('+1 more')).not.toBeInTheDocument();
});
