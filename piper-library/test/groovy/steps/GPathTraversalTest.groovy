package steps

import hudson.AbortException
import net.sf.json.JSONException
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.RuleChain
import util.*

import static org.hamcrest.Matchers.containsString
import static org.hamcrest.Matchers.is
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertTrue

class GPathTraversalTest {

    @Test
    void testReadVersionUseCaseLayerUp() {
        def object = [a : [ b: [ c: 42 ] ] ]
        def test = "a.b.c"
        def result = test.split(/\./).inject(null) { curr, prop ->
            curr?."$prop" ?: object[prop]
        }
        println(result)
    }
}
