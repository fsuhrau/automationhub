import {PlatformType} from "../types/platform.type.enum";
import {
    Public
} from "@mui/icons-material";
import React from "react";
import {ReactComponent as UnityIcon} from "../icons/unity-svgrepo-com.svg";
import {ReactComponent as MacOSIcon} from "../icons/macos-svgrepo-com-2.svg";
import {ReactComponent as IOSIcon} from "../icons/ios-svgrepo-com.svg";
import {ReactComponent as AndroidIcon} from "../icons/android-svgrepo-com.svg";
import {ReactComponent as LinuxIcon} from "../icons/linux-svgrepo-com.svg";
import {ReactComponent as WindowsIcon} from "../icons/windows-174-svgrepo-com.svg";

import {useTheme} from "@mui/material/styles";

interface PlatformTypeIconProps {
    platformType: PlatformType
}

const PlatformTypeIcon: React.FC<PlatformTypeIconProps> = (props: PlatformTypeIconProps) => {

    const {platformType} = props;
    const theme = useTheme();

    return (<>
        {platformType === PlatformType.iOS && <IOSIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Android && <AndroidIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Mac && <MacOSIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Windows && <WindowsIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Linux && <LinuxIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Web && <Public style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Editor && <UnityIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }} />}
    </>)
}

export default PlatformTypeIcon;