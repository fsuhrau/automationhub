import {PlatformType} from "../types/platform.type.enum";
import {AndroidRounded, Apple} from "@mui/icons-material";
import React from "react";
import DevicesRoundedIcon from "@mui/icons-material/DevicesRounded";
import ConstructionRoundedIcon from "@mui/icons-material/ConstructionRounded";

interface PlatformTypeIconProps {
    platformType: PlatformType
}

const PlatformTypeIcon: React.FC<PlatformTypeIconProps> = (props: PlatformTypeIconProps) => {

    const {platformType} = props;

    return (<>
        {platformType === PlatformType.iOS && <Apple/>}
        {platformType === PlatformType.Android && <AndroidRounded/>}
        {platformType === PlatformType.Mac && <ConstructionRoundedIcon/>}
        {platformType === PlatformType.Windows && <ConstructionRoundedIcon/>}
        {platformType === PlatformType.Linux && <ConstructionRoundedIcon/>}
        {platformType === PlatformType.Web && <ConstructionRoundedIcon/>}
        {platformType === PlatformType.Editor && <DevicesRoundedIcon/>}
    </>)
}

export default PlatformTypeIcon;