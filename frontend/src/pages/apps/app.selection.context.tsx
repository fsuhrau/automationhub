import React, { useEffect } from "react";
import { ApplicationProps } from "../../application/application.props";
import { useProjectAppContext } from "../../project/app.context";
import { useNavigate } from "react-router-dom";
import { SelectChangeEvent } from "@mui/material";

interface AppSelectionProps extends ApplicationProps {
    path: string
    children?: React.ReactNode
}

const AppSelectionContext: React.FC<AppSelectionProps> = (props) => {

    const {path, children, appState, dispatch} = props;

    const {projectId, appId} = useProjectAppContext();

    const navigate = useNavigate();

    useEffect(() => {
        if (appId === 0) {
            const value = appState.project?.Apps === undefined ? null : appState.project?.Apps.length === 0 ? null : appState.project?.Apps[ 0 ].ID;
            if (value !== null) {
                navigate(`/project/${ projectId }/app/${ value }/${ path }`)
            }
        }
        else {

        }
    }, [appId, appState.project?.Apps])

    return (<>{children}</>);
};

export default AppSelectionContext;
