import {PlatformType} from "../types/platform.type.enum";
import {Android, Apple, Web} from "@mui/icons-material";
import React from "react";
import ConstructionRoundedIcon from "@mui/icons-material/ConstructionRounded";
import {ReactComponent as UnityIcon} from "../icons/unity-svgrepo-com.svg";
import {useTheme} from "@mui/material/styles";

interface PlatformTypeIconProps {
    platformType: PlatformType
}

const PlatformTypeIcon: React.FC<PlatformTypeIconProps> = (props: PlatformTypeIconProps) => {

    const {platformType} = props;
    const theme = useTheme();

    return (<>
        {platformType === PlatformType.iOS && <Apple style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Android && <Android style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Mac && <ConstructionRoundedIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Windows && <ConstructionRoundedIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Linux && <ConstructionRoundedIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Web && <Web style={{ fill: theme.palette.text.primary, width: 24, height: 24 }}/>}
        {platformType === PlatformType.Editor && <UnityIcon style={{ fill: theme.palette.text.primary, width: 24, height: 24 }} />}
    </>)
}

export default PlatformTypeIcon;