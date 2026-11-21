import * as React from 'react';
import {styled} from '@mui/material/styles';
import Breadcrumbs, {breadcrumbsClasses} from '@mui/material/Breadcrumbs';
import Link from '@mui/material/Link';
import Typography from '@mui/material/Typography';
import NavigateNextRoundedIcon from '@mui/icons-material/NavigateNextRounded';
import {Link as RouterLink, useLocation} from 'react-router-dom';
import {useHubState} from '../hooks/HubStateProvider';
import {useTestService} from '../services/test.service';
import {buildCrumbs, Crumb, CrumbNames, parseTestPath} from './breadcrumbs';
import {HubState} from '../application/HubState';

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

const testKey = (project: string, appId: number, testId: number | string): string => `${project}/${appId}/${testId}`;

const useTestName = (pathname: string): void => {
    const [, forceUpdate] = React.useReducer((x: number) => x + 1, 0);
    const service = useTestService();
    const target = parseTestPath(pathname);
    const project = target?.project;
    const appId = target?.appId;
    const testId = target?.testId;
    React.useEffect(() => {
        if (project === undefined || appId === undefined || testId === undefined) return;
        const key = testKey(project, appId, testId);
        if (testNames.has(key)) return;
        let active = true;
        service.get({projectId: project, appId}, String(testId))
            .then(test => {
                testNames.set(key, test.name);
                if (active) forceUpdate();
            })
            .catch(() => undefined);
        return () => {
            active = false;
        };
    }, [project, appId, testId, service]);
};

const deviceName = (state: HubState, id: number): string | undefined => {
    const d = state.devices?.find(device => device.id === id);
    if (!d) return undefined;
    return d.alias?.length > 0 ? d.alias : d.name;
};

const namesOf = (state: HubState, pathname: string): CrumbNames => ({
    project: id => state.projects?.find(p => p.identifier === id)?.name,
    app: id => state.apps?.find(a => a.id === id)?.name,
    device: id => deviceName(state, id),
    test: (appId, id) => {
        const target = parseTestPath(pathname);
        return target ? testNames.get(testKey(target.project, appId, id)) : undefined;
    },
});

const CrumbItem: React.FC<{ crumb: Crumb }> = ({crumb}) => crumb.to
    ? <Link component={RouterLink} to={crumb.to} underline="hover" color="text.secondary" variant="body1">{crumb.label}</Link>
    : <Typography variant="body1" aria-current="page" sx={{color: 'text.primary', fontWeight: 600}}>{crumb.label}</Typography>;

export default function NavbarBreadcrumbs(): React.ReactElement {
    const {pathname} = useLocation();
    const {state} = useHubState();
    useTestName(pathname);
    const crumbs = buildCrumbs(pathname, namesOf(state, pathname));

    return (
        <StyledBreadcrumbs aria-label="breadcrumb" separator={<NavigateNextRoundedIcon fontSize="small"/>}>
            {crumbs.map(c => <CrumbItem key={c.label + (c.to ?? '')} crumb={c}/>)}
        </StyledBreadcrumbs>
    );
}
