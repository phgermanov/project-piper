package com.sap.piper.internal

import static org.hamcrest.Matchers.*

import static org.junit.Assert.assertThat
import org.junit.Test


class BashUtilsTest {
    @Test
    void testNoSingleQuote() {
        assertThat(BashUtils.escape('ZdTq8@5gj$^9yYMy'), is('\'ZdTq8@5gj$^9yYMy\''))
    }

    @Test
    void testWithSingleQuote() {
        assertThat(BashUtils.escape("ZdTq8@5gj'^9yYMy"), is('\'ZdTq8@5gj\'"\'"\'^9yYMy\''))
    }

}
