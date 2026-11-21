import {expect, it} from 'vitest';
import {render, screen} from '@testing-library/react';
import {PlatformType} from '../../../types/platform.type.enum';
import {PlatformNotice} from './PlatformNotice';

it.each([
    [PlatformType.Editor, 'Editor apps don\'t need a bundle: tests run in the Unity editor.'],
    [PlatformType.Web, 'Web apps don\'t need a bundle: tests run in the browser.'],
])('tells that an app of platform %s needs no bundle', (platform, text) => {
    render(<PlatformNotice platform={platform}/>);
    expect(screen.getByText(text)).toBeInTheDocument();
});

it.each([PlatformType.Android, undefined])('says nothing for platform %s', platform => {
    const {container} = render(<PlatformNotice platform={platform}/>);
    expect(container).toBeEmptyDOMElement();
});
