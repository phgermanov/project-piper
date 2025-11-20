package com.sap.piper.internal


import hudson.plugins.sidebar_link.LinkAction
import hudson.triggers.TimerTrigger
import org.jenkinsci.plugins.workflow.libs.LibrariesAction
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.rules.ExpectedException
import org.junit.rules.RuleChain
import util.BasePiperTest
import util.JenkinsEnvironmentRule
import util.JenkinsLoggingRule
import util.JenkinsReadFileRule
import util.LibraryLoadingTestExecutionListener
import util.Rules

import static org.hamcrest.Matchers.allOf
import static org.hamcrest.Matchers.hasEntry
import static org.hamcrest.Matchers.hasKey
import static org.hamcrest.Matchers.is
import static org.hamcrest.Matchers.isEmptyOrNullString
import static org.junit.Assert.assertThat
import static org.junit.Assert.assertEquals


class JenkinsUtilsTest extends BasePiperTest {

    public ExpectedException exception = ExpectedException.none()
    public JenkinsLoggingRule loggingRule = new JenkinsLoggingRule(this)
    private JenkinsEnvironmentRule jer = new JenkinsEnvironmentRule(this)
    JenkinsReadFileRule jrfr = new JenkinsReadFileRule(this, 'test/resources/jenkinsUtilsTest')

    @Rule
    public RuleChain ruleChain = Rules.getCommonRules(this)
        .around(exception)
        .around(loggingRule)
        .around(jer)
        .around(jrfr)

    JenkinsUtils jenkinsUtils
    Object currentBuildMock
    Object rawBuildMock
    Object jenkinsInstanceMock
    Object parentMock

    Map triggerCause


    @Before
    void init() throws Exception {
        jenkinsUtils = new JenkinsUtils() {
            def getCurrentBuildInstance() {
                return currentBuildMock
            }

            def getActiveJenkinsInstance() {
                return jenkinsInstanceMock
            }
        }
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(jenkinsUtils)

        jenkinsInstanceMock = new Object()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(jenkinsInstanceMock)

        parentMock = new Object() {

        }
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(parentMock)

        rawBuildMock = new Object() {
            def getParent() {
                return parentMock
            }
            def getCause(type) {
                return triggerCause
            }

        }
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(rawBuildMock)

        currentBuildMock = new Object() {
            def number
            def getRawBuild() {
                return rawBuildMock
            }
        }
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(currentBuildMock)
    }

    @Test
    void testAddBuildDiscarder() {
        def discarderObject
        helper.registerAllowedMethod("getBuildDiscarder", [], {
            return null
        })
        helper.registerAllowedMethod("setBuildDiscarder", [hudson.tasks.LogRotator.class], {
            newDiscarder -> discarderObject = newDiscarder
        })

        jenkinsUtils.addBuildDiscarder(-1, 35)

        assertEquals(-1, discarderObject.getDaysToKeep())
        assertEquals(35, discarderObject.getNumToKeep())
        assertEquals(-1, discarderObject.getArtifactDaysToKeep())
        assertEquals(-1, discarderObject.getArtifactNumToKeep())
    }

    @Test
    void testAddBuildDiscarderFullSpec() {
        def discarderObject
        helper.registerAllowedMethod("getBuildDiscarder", [], {
            return null
        })
        helper.registerAllowedMethod("setBuildDiscarder", [hudson.tasks.LogRotator.class], {
            newDiscarder -> discarderObject = newDiscarder
        })

        helper.registerAllowedMethod("addBuildDiscarder", [int, int, int, int], null)

        jenkinsUtils.addBuildDiscarder(5, 35, 15, 2)

        assertEquals(5, discarderObject.getDaysToKeep())
        assertEquals(35, discarderObject.getNumToKeep())
        assertEquals(15, discarderObject.getArtifactDaysToKeep())
        assertEquals(2, discarderObject.getArtifactNumToKeep())
    }

    @Test
    void testRemoveBuildDiscarder() {
        def discarderObject = new hudson.tasks.LogRotator(-1, 30, -1, -1)

        helper.registerAllowedMethod("getBuildDiscarder", [], {
            return discarderObject
        })
        helper.registerAllowedMethod("setBuildDiscarder", [hudson.tasks.LogRotator.class], {
            newDiscarder -> discarderObject = newDiscarder
        })

        jenkinsUtils.removeBuildDiscarder()

        assertEquals(null, discarderObject)
    }

    @Test
    void testAddGlobalSideBarLink() {
        def actions = new ArrayList()

        helper.registerAllowedMethod("getActions", [], {
            return actions
        })

        jenkinsUtils.addGlobalSideBarLink("abcd/1234", "Some report link", "images/24x24/report.png")

        assertEquals(1, actions.size())
        assertEquals(LinkAction.class, actions.get(0).getClass())
        assertEquals("abcd/1234", actions.get(0).getUrlName())
        assertEquals("Some report link", actions.get(0).getDisplayName())
        assertEquals("/images/24x24/report.png", actions.get(0).getIconFileName())
    }

    @Test
    void testRemoveGlobalSideBarLinks() {
        def actions = new ArrayList()
        actions.add(new LinkAction("abcd/1234", "Some report link", "images/24x24/report.png"))

        helper.registerAllowedMethod("getActions", [], {
            return actions
        })

        jenkinsUtils.removeGlobalSideBarLinks("abcd/1234")

        assertEquals(0, actions.size())
    }

    @Test
    void testAddJobSideBarLink() {
        def actions = new ArrayList()

        helper.registerAllowedMethod("getActions", [], {
            return actions
        })

        currentBuildMock.number = 15

        jenkinsUtils.addJobSideBarLink("abcd/1234", "Some report link", "images/24x24/report.png")

        assertEquals(1, actions.size())
        assertEquals(LinkAction.class, actions.get(0).getClass())
        assertEquals("15/abcd/1234", actions.get(0).getUrlName())
        assertEquals("Some report link", actions.get(0).getDisplayName())
        assertEquals("/images/24x24/report.png", actions.get(0).getIconFileName())
    }

    @Test
    void testRemoveJobSideBarLinks() {
        def actions = new ArrayList()
        actions.add(new LinkAction("abcd/1234", "Some report link", "images/24x24/report.png"))

        helper.registerAllowedMethod("getActions", [], {
            return actions
        })

        jenkinsUtils.removeJobSideBarLinks("abcd/1234")

        assertEquals(0, actions.size())
    }

    @Test
    void testAddRunSideBarLink() {
        def actions = new ArrayList()

        helper.registerAllowedMethod("getActions", [], {
            return actions
        })

        jenkinsUtils.addRunSideBarLink("abcd/1234", "Some report link", "images/24x24/report.png")

        assertEquals(1, actions.size())
        assertEquals(LinkAction.class, actions.get(0).getClass())
        assertEquals("abcd/1234", actions.get(0).getUrlName())
        assertEquals("Some report link", actions.get(0).getDisplayName())
        assertEquals("/images/24x24/report.png", actions.get(0).getIconFileName())
    }


    @Test
    void testIsJobStartedByTimer() {
        helper.registerAllowedMethod("getCause", [hudson.triggers.TimerTrigger.TimerTriggerCause.class], {
            return new hudson.triggers.TimerTrigger.TimerTriggerCause()
        })

        def result = jenkinsUtils.isJobStartedByTimer()

        assertEquals(true, result)
    }

    @Test
    void testIsJobStartedByUser() {
        helper.registerAllowedMethod("getCause", [hudson.model.Cause.UserIdCause.class], {
            return new hudson.model.Cause.UserIdCause()
        })

        def result = jenkinsUtils.isJobStartedByUser()

        assertEquals(true, result)
    }

    @Test
    void testIsJobNotStartedByTimer() {
        helper.registerAllowedMethod("getCause", [hudson.triggers.TimerTrigger$TimerTriggerCause], {
            return null
        })

        def result = jenkinsUtils.isJobStartedByTimer()

        assertEquals(false, result)
    }

    @Test
    void testNewScheduleJob() {
        def triggers = new ArrayList()
        def jobPropertyMock = new Object()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(jobPropertyMock)
        helper.registerAllowedMethod("getTriggersJobProperty", [], {
            return jobPropertyMock
        })
        helper.registerAllowedMethod("getTriggers", [], {
            return triggers
        })
        helper.registerAllowedMethod("setTriggers", [ArrayList.class], {
            newTriggers -> triggers = newTriggers
        })

        jenkinsUtils.scheduleJob("* * 15 * *")

        assertEquals(1, triggers.size())
        assertEquals(TimerTrigger.class, triggers.get(0).getClass())
        assertEquals("* * 15 * *", triggers.get(0).getSpec())
    }

    @Test
    void testAdditionalScheduleJob() {
        def triggers = new ArrayList()
        triggers.add(new hudson.triggers.TimerTrigger("* * 1 2 3"))
        def jobPropertyMock = new Object()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(jobPropertyMock)
        helper.registerAllowedMethod("getTriggersJobProperty", [], {
            return jobPropertyMock
        })
        helper.registerAllowedMethod("getTriggers", [], {
            return triggers
        })
        helper.registerAllowedMethod("setTriggers", [ArrayList.class], {
            newTriggers -> triggers = newTriggers
        })

        jenkinsUtils.scheduleJob("* * 15 * *")

        assertEquals(1, triggers.size())
        assertEquals(TimerTrigger.class, triggers.get(0).getClass())
        assertEquals("* * 1 2 3\n* * 15 * *", triggers.get(0).getSpec())
    }

    @Test
    void testRemoveFirstJobSchedule() {
        def triggers = new ArrayList()
        triggers.add(new hudson.triggers.TimerTrigger("* * 1 2 3\n* * 15 * *"))
        def jobPropertyMock = new Object()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(jobPropertyMock)
        helper.registerAllowedMethod("getTriggersJobProperty", [], {
            return jobPropertyMock
        })
        helper.registerAllowedMethod("getTriggers", [], {
            return triggers
        })
        helper.registerAllowedMethod("setTriggers", [ArrayList.class], {
            newTriggers -> triggers = newTriggers
        })

        jenkinsUtils.removeJobSchedule("* * 1 2 3")

        assertEquals(1, triggers.size())
        assertEquals(TimerTrigger.class, triggers.get(0).getClass())
        assertEquals("* * 15 * *", triggers.get(0).getSpec())
    }

    @Test
    void testRemoveSecondJobSchedule() {
        def triggers = new ArrayList()
        triggers.add(new hudson.triggers.TimerTrigger("* * 1 2 3\n* * 15 * *"))
        def jobPropertyMock = new Object()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(jobPropertyMock)
        helper.registerAllowedMethod("getTriggersJobProperty", [], {
            return jobPropertyMock
        })
        helper.registerAllowedMethod("getTriggers", [], {
            return triggers
        })
        helper.registerAllowedMethod("setTriggers", [ArrayList.class], {
            newTriggers -> triggers = newTriggers
        })

        jenkinsUtils.removeJobSchedule("* * 15 * *")

        assertEquals(1, triggers.size())
        assertEquals(TimerTrigger.class, triggers.get(0).getClass())
        assertEquals("* * 1 2 3", triggers.get(0).getSpec())
    }

    @Test
    void testRemoveAllJobSchedules() {
        def triggers = new ArrayList()
        triggers.add(new hudson.triggers.TimerTrigger("* * 1 2 3\n* * 15 * *"))
        def jobPropertyMock = new Object()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(jobPropertyMock)
        helper.registerAllowedMethod("getTriggersJobProperty", [], {
            return jobPropertyMock
        })
        helper.registerAllowedMethod("getTriggers", [], {
            return triggers
        })
        helper.registerAllowedMethod("setTriggers", [ArrayList.class], {
            newTriggers -> triggers = newTriggers
        })

        jenkinsUtils.removeJobSchedule()

        assertEquals(0, triggers.size())
    }

    @Test
    void testGetLibrariesInfoWithPiperLatest() {
        def libraries = new ArrayList()
        libraries.add([name: 'a', version: '1', trusted: true])
        libraries.add([name: 'b', version: '2', trusted: true])
        def librariesActionMock = new Object() {
            def getLibraries() {
                return libraries
            }
        }
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(librariesActionMock)
        helper.registerAllowedMethod("getAction", [LibrariesAction.class], {
            return librariesActionMock
        })
        helper.registerAllowedMethod("sh", [String.class], {
            return 0
        })
        helper.registerAllowedMethod("sh", [LinkedHashMap.class], {
            return '1.2.3-31273627836+3562715386512eghfef12f6725e master'
        })

        def result = jenkinsUtils.getLibrariesInfoWithPiperLatest()

        assertEquals(3, result.size())
    }

    @Test
    void testGetIssueCommentTriggerAction() {
        triggerCause = [
            comment: 'this is my test comment /n /piper test whatever',
            triggerPattern: '.*/piper ([a-z]*).*'
        ]
        assertThat(jenkinsUtils.getIssueCommentTriggerAction(), is('test'))
    }

    @Test
    void testGetIssueCommentTriggerActionNoAction() {
        triggerCause = [
            comment: 'this is my test comment /n whatever',
            triggerPattern: '.*/piper ([a-z]*).*'
        ]
        assertThat(jenkinsUtils.getIssueCommentTriggerAction(), isEmptyOrNullString())
    }

    @Test
    void testGetPlugin() {
        def testObject = new Object() {
            def getShortName(){
                return "Test Plugin Object"
            }
        }
        helper.registerAllowedMethod('getPluginManager', [], { return new Object() {
            def getPlugins(){
                return [testObject]
            }
        }})
        assertThat(jenkinsUtils.getPlugin("Test Plugin Object"), is(testObject))
        assertThat(jenkinsUtils.getPlugin("Other Plugin Object"), is(null))
    }

    @Test
    void testGetCheckmarxResults() {
        helper.registerAllowedMethod('findFiles', [LinkedHashMap], {
            return [new File('checkmarx/ScanReport.xml')]
        })

        def sastResults = jenkinsUtils.getCheckmarxResults()

        assertThat(sastResults, allOf(hasKey('High'), hasKey('Medium'), hasKey('Low'), hasKey('Information')))
        assertThat(sastResults.High, allOf(hasEntry('Issues', 0), hasEntry('NotFalsePositive', 0), hasEntry('ToVerify', 0), hasEntry('Confirmed', 0), hasEntry('NotExploitable', 0), hasEntry('Urgent', 0), hasEntry('ProposedNotExploitable', 0)))
        assertThat(sastResults.Medium, allOf(hasEntry('Issues', 1), hasEntry('NotFalsePositive', 0), hasEntry('ToVerify', 0), hasEntry('Confirmed', 0), hasEntry('NotExploitable', 1), hasEntry('Urgent', 0), hasEntry('ProposedNotExploitable', 0)))
        assertThat(sastResults.Low, allOf(hasEntry('Issues', 2), hasEntry('NotFalsePositive', 2), hasEntry('ToVerify', 1), hasEntry('Confirmed', 1), hasEntry('NotExploitable', 0), hasEntry('Urgent', 0), hasEntry('ProposedNotExploitable', 0)))
        assertThat(sastResults.Information, allOf(hasEntry('Issues', 3), hasEntry('NotFalsePositive', 3), hasEntry('ToVerify', 1), hasEntry('Confirmed', 0), hasEntry('NotExploitable', 0), hasEntry('Urgent', 1), hasEntry('ProposedNotExploitable', 1)))
    }

}
