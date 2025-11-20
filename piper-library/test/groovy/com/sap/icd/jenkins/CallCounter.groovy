package com.sap.icd.jenkins

import java.util.concurrent.atomic.AtomicInteger

class CallCounter {
    private AtomicInteger calls = new AtomicInteger(0)

    public void called() {
        calls.incrementAndGet()
    }

    public boolean gotCalled(int called) {
        return calls.get() == called
    }

    public int getCalls() {
        return calls.get()
    }

    @Override
    public String toString() {
        return calls.toString()
    }
}
