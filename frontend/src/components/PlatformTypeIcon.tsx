import React from "react";
import {Public} from "@mui/icons-material";
import SvgIcon, {SvgIconProps} from "@mui/material/SvgIcon";
import {PlatformType} from "../types/platform.type.enum";
import UnityIcon from "../icons/unity-svgrepo-com.svg?react";
import MacOSIcon from "../icons/macos-svgrepo-com-2.svg?react";
import IOSIcon from "../icons/ios-svgrepo-com.svg?react";
import AndroidIcon from "../icons/android-svgrepo-com.svg?react";
import LinuxIcon from "../icons/linux-svgrepo-com.svg?react";
import WindowsIcon from "../icons/windows-174-svgrepo-com.svg?react";

const icons: Partial<Record<PlatformType, React.ElementType>> = {
    [PlatformType.iOS]: IOSIcon,
    [PlatformType.IOSSimulator]: IOSIcon,
    [PlatformType.Android]: AndroidIcon,
    [PlatformType.Mac]: MacOSIcon,
    [PlatformType.Windows]: WindowsIcon,
    [PlatformType.Linux]: LinuxIcon,
    [PlatformType.Editor]: UnityIcon,
};

interface PlatformTypeIconProps extends SvgIconProps {
    platformType: PlatformType
}

// PlatformTypeIcon renders as a MUI icon, so it uses the current text color
// (light and dark mode) and the usual icon sizes.
const PlatformTypeIcon: React.FC<PlatformTypeIconProps> = ({platformType, ...props}) => {
    if (platformType === PlatformType.Web) {
        return <Public {...props}/>;
    }
    const icon = icons[platformType];
    if (!icon) {
        return null;
    }
    return <SvgIcon component={icon} inheritViewBox {...props}/>;
};

export default PlatformTypeIcon;
