import React from 'react';
import {Box, Button, Typography} from '@mui/material';
import {Link as RouterLink} from 'react-router-dom';
import {Section} from '../../../../components/common';
import ITestConfigData from '../../../../types/test.config';
import ITestConfigUnityData from '../../../../types/test.config.unity';
import {TestType} from '../../../../types/test.type.enum';
import {UnityTestCategory} from '../../../../types/unity.test.category.type.enum';
import {methodsByClass} from '../testShowModel';

const Headline: React.FC<{ unity: ITestConfigUnityData }> = ({unity}) => {
    switch (unity.testCategoryType) {
        case UnityTestCategory.RunAllTests:
            return <>All Unity tests of the app.</>;
        case UnityTestCategory.RunAllOfCategory:
            return <>All Unity tests of {unity.categories ? <>the categories <b>{unity.categories}</b></> : 'a category (none set)'}.</>;
        default:
            return <>{unity.testFunctions.length} picked Unity {unity.testFunctions.length === 1 ? 'test' : 'tests'}:</>;
    }
};

const PickedTests: React.FC<{ unity: ITestConfigUnityData }> = ({unity}) =>
    <Box sx={{maxHeight: 320, overflowY: 'auto'}}>
        {methodsByClass(unity.testFunctions).map(([cls, methods]) => <Box key={cls} sx={{mb: 1}}>
            <Typography variant="caption" color="text.secondary" sx={{wordBreak: 'break-all'}}>{cls}</Typography>
            {methods.map(m => <Typography key={m} variant="body2" sx={{fontFamily: 'monospace', fontSize: 13, pl: 1}}>{m}</Typography>)}
        </Box>)}
    </Box>;

const UnityTests: React.FC<{ unity: ITestConfigUnityData }> = ({unity}) =>
    <>
        <Typography variant="body2" sx={{mb: 1}}><Headline unity={unity}/></Typography>
        {unity.testCategoryType === UnityTestCategory.RunSelectedTestsOnly && <PickedTests unity={unity}/>}
    </>;

// WhatItRunsCard says which tests the test starts
export const WhatItRunsCard: React.FC<{ config: ITestConfigData, editUrl: string }> = ({config, editUrl}) => {
    const unity = config.type === TestType.Unity ? config.unity : null;
    return <Section title="What it runs" action={<Button size="small" component={RouterLink} to={editUrl}>Edit</Button>}>
        {unity
            ? <UnityTests unity={unity}/>
            : <Typography variant="body2">A {TestType[config.type] ?? ''} test.</Typography>}
    </Section>;
};
