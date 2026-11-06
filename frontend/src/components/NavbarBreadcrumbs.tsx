import * as React from 'react';
import {styled} from '@mui/material/styles';
import Breadcrumbs, {breadcrumbsClasses} from '@mui/material/Breadcrumbs';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import NavigateNextRoundedIcon from '@mui/icons-material/NavigateNextRounded';
import {Link as RouterLink, useLocation} from 'react-router-dom';
import {useHubState} from '../hooks/HubStateProvider';
import {getTest} from '../services/test.service';
import {buildCrumbs} from './breadcrumbs';

const StyledBreadcrumbs = styled(Breadcrumbs)(({theme}) => ({
    margin: theme.spacing(1, 0),
    [`& .${breadcrumbsClasses.separator}`]: {
        color: theme.palette.action.disabled,
        margin: 1,
    },
    [`& .${breadcrumbsClasses.ol}`]: {
        alignItems: 'center',
    },
}));

// test names aren't in the hub state; they are looked up once per test
const testNames = new Map<string, string>();

const useTestName = (pathname: string): void => {
    const [, forceUpdate] = React.useReducer((x: number) => x + 1, 0);
    const match = pathname.match(/^\/project\/([^/]+)\/app:(\d+)\/test\/(\d+)/);
    const key = match ? `${match[1]}/${match[2]}/${match[3]}` : null;
    React.useEffect(() => {
        if (!match || !key || testNames.has(key)) return;
        let active = true;
        getTest(decodeURIComponent(match[1]), +match[2], match[3])
            .then(test => {
                testNames.set(key, test.name);
                if (active) forceUpdate();
            })
            .catch(() => undefined);
        return () => {
            active = false;
        };
    }, [key]);
};

export default function NavbarBreadcrumbs() {
    const {pathname} = useLocation();
    const {state} = useHubState();
    useTestName(pathname);

    const crumbs = buildCrumbs(pathname, {
        project: id => state.projects?.find(p => p.identifier === id)?.name,
        app: id => state.apps?.find(a => a.id === id)?.name,
        device: id => {
            const d = state.devices?.find(d => d.id === id);
            return d ? (d.alias?.length > 0 ? d.alias : d.name) : undefined;
        },
        test: (appId, id) => {
            const project = pathname.split('/')[2];
            return testNames.get(`${project}/${appId}/${id}`);
        },
    });

    return (
        <StyledBreadcrumbs
            aria-label="breadcrumb"
            separator={<NavigateNextRoundedIcon fontSize="small"/>}
        >
            {crumbs.map((c, i) => c.to
                ? <Link key={i} component={RouterLink} to={c.to} underline="hover" color="text.secondary"
                        variant="body1">{c.label}</Link>
                : <Typography key={i} variant="body1" aria-current="page"
                              sx={{color: 'text.primary', fontWeight: 600}}>{c.label}</Typography>)}
        </StyledBreadcrumbs>
    );
}
