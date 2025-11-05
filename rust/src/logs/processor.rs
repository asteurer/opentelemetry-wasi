use crate::wit::wasi;
use opentelemetry_sdk::error::OTelSdkResult;
use std::sync::atomic::{AtomicBool, Ordering};

#[derive(Debug)]
pub struct WasiLogProcessor {
    is_shutdown: AtomicBool,
}

impl WasiLogProcessor {
    pub fn new() -> Self {
        Self {
            is_shutdown: AtomicBool::new(false),
        }
    }
}

impl Default for WasiLogProcessor {
    fn default() -> Self {
        Self::new()
    }
}

impl opentelemetry_sdk::logs::LogProcessor for WasiLogProcessor {
    fn emit(
        &self,
        data: &mut opentelemetry_sdk::logs::SdkLogRecord,
        _: &opentelemetry::InstrumentationScope,
    ) {
        if let Err(e) = wasi::otel::logs::emit(&data.into()) {
            opentelemetry::otel_error!(name: "emit_error", msg = e);
        }
    }

    fn force_flush(&self) -> opentelemetry_sdk::error::OTelSdkResult {
        if self.is_shutdown.load(Ordering::Relaxed) {
            return OTelSdkResult::Err(opentelemetry_sdk::error::OTelSdkError::AlreadyShutdown);
        }
        Ok(())
    }

    fn shutdown(&self) -> opentelemetry_sdk::error::OTelSdkResult {
        let result = self.force_flush();
        if self.is_shutdown.swap(true, Ordering::Relaxed) {
            return OTelSdkResult::Err(opentelemetry_sdk::error::OTelSdkError::AlreadyShutdown);
        }
        result
    }
}
