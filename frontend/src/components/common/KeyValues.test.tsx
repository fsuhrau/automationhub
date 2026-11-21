import {describe, expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';
import {KeyValues} from './KeyValues';

describe('KeyValues', () => {
    it('lists names and values', () => {
        render(<KeyValues rows={[['OS', 'Android'], ['Version', <b key="v">14</b>]]}/>);
        expect(screen.getByText('OS')).toBeInTheDocument();
        expect(screen.getByText('Android')).toBeInTheDocument();
        expect(screen.getByText('14')).toBeInTheDocument();
    });

    it('can set the values in monospace', () => {
        render(<KeyValues monospace={true} rows={[['Hash', 'abc']]}/>);
        expect(screen.getByText('abc')).toHaveStyle({fontFamily: 'monospace'});
    });
});
