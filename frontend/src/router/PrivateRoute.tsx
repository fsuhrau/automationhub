import React from "react";
import {Navigate, Outlet} from "react-router-dom";
import {Backdrop, CircularProgress} from "@mui/material";
import {useAuth} from "../hooks/AuthProvider";

const PrivateRoute = () => {
    const {authState} = useAuth();
    if (authState === 'checking') {
        return <Backdrop sx={{zIndex: 1, color: '#fff'}} open={true}><CircularProgress color="inherit"/></Backdrop>;
    }
    if (authState === 'anonymous') {
        return <Navigate to="/login"/>;
    }
    return <Outlet/>;
};

export default PrivateRoute;
