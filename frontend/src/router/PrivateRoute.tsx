import React from "react";
import {Navigate, Outlet, useLocation} from "react-router-dom";
import {Backdrop, CircularProgress} from "@mui/material";
import {useAuth} from "../hooks/AuthProvider";
import {routes} from "../lib/routes";

const PrivateRoute = (): React.ReactElement => {
    const {authState} = useAuth();
    const location = useLocation();
    if (authState === 'checking') {
        return <Backdrop sx={{zIndex: 1, color: '#fff'}} open={true}><CircularProgress color="inherit"/></Backdrop>;
    }
    if (authState === 'anonymous') {
        // after signing in, the sign in page returns here
        return <Navigate to={routes.login()} replace={true} state={{from: location.pathname + location.search}}/>;
    }
    return <Outlet/>;
};

export default PrivateRoute;
