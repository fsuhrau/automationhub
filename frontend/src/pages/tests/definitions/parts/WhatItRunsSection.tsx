import React from 'react';
import {Box, RadioGroup} from '@mui/material';
import {Section} from '../../../../components/common';
import {UnityTestCategory} from '../../../../types/unity.test.category.type.enum';
import {FormErrors, TestForm} from '../../testEditModel';
import {CategoryEditor} from './CategoryEditor';
import {Choice} from './Choice';
import {TestPicker} from './TestPicker';

interface WhatItRunsSectionProps {
    form: TestForm,
    set: (change: Partial<TestForm>) => void,
    // the errors to show, none before a save was tried
    errors: FormErrors,
}

// WhatItRunsSection picks which Unity tests the test starts
export const WhatItRunsSection: React.FC<WhatItRunsSectionProps> = ({form, set, errors}) =>
    <Section title="What it runs" description="Which Unity tests of the app the test starts.">
        <RadioGroup value={String(form.mode)} onChange={e => set({mode: Number(e.target.value) as UnityTestCategory})}>
            <Choice value={String(UnityTestCategory.RunAllTests)} label="All tests" hint="Every Unity test the app has, including new ones."/>
            <Choice value={String(UnityTestCategory.RunAllOfCategory)} label="Tests of categories" hint="The Unity tests with one of the named categories."/>
            <Choice value={String(UnityTestCategory.RunSelectedTestsOnly)} label="Picked tests" hint="Only the tests picked here."/>
        </RadioGroup>
        {form.mode === UnityTestCategory.RunAllOfCategory &&
            <CategoryEditor categories={form.categories} onChange={categories => set({categories})} error={errors.categories}/>}
        {form.mode === UnityTestCategory.RunSelectedTestsOnly && <Box sx={{mt: 1.5}}>
            <TestPicker picked={form.functions} onChange={functions => set({functions})} error={errors.tests}/>
        </Box>}
    </Section>;
