import React from "react";
import { Outlet, useNavigate, useParams } from "react-router-dom";
import { ProjectAppProvider } from "../../project/app.context";

const ProjectAppPage: React.FC = () => {
    const { project_id, app_id } = useParams();
    const navigate = useNavigate();
    return (project_id !== undefined && app_id !== undefined ? <ProjectAppProvider><Outlet /></ProjectAppProvider> : null);
};

export default ProjectAppPage;