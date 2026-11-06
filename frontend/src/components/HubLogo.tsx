import * as React from 'react';
import SvgIcon, {SvgIconProps} from '@mui/material/SvgIcon';
import HubIcon from '../icons/tech-icon.svg?react';

// HubLogo is the Automation Hub mark (the gear of the favicon), in the theme's primary color
export default function HubLogo(props: SvgIconProps) {
    return <SvgIcon component={HubIcon} inheritViewBox aria-hidden {...props}
                    sx={[{color: 'primary.main'}, ...(Array.isArray(props.sx) ? props.sx : [props.sx])]}/>;
}
