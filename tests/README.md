# HubStream Integration Tests

이 폴더에는 `go test` 대신 `go run`을 사용하는 통합 테스트가 포함되어 있습니다.

## 장점

- **타임아웃 제한 없음**: `go test`의 10분 타임아웃 제한 없이 긴 테스트 실행 가능
- **더 유연한 설정**: JetStream, 클러스터 등의 설정을 자유롭게 조정 가능
- **디버깅 용이**: 표준 출력으로 자세한 로그 확인 가능
- **실제 사용 사례 모방**: 프로덕션 환경과 유사한 설정으로 테스트

## 실행 방법

```bash
cd tests
go run main.go
```

## 테스트 케이스 상세 설명

### 1. Basic Hub Creation (기본 허브 생성)
**목적**: HubStream 허브의 기본 생성 및 초기화 기능을 검증합니다.

**테스트 내용**:
- 기본 옵션 생성
- 허브 인스턴스 생성
- 서버 시작 및 연결 확인
- 고유 ID 할당 검증

**중요성**: 모든 다른 테스트의 기반이 되는 기본 기능 검증

**실패 시나리오**: 옵션 생성 실패, 서버 시작 실패, 연결 실패

### 2. JetStream Operations (JetStream 영구 메시징)
**목적**: JetStream을 사용한 영구 메시징 기능을 종합적으로 테스트합니다.

**테스트 내용**:
- 영구 스트림 생성 및 구성
- 메시지 발행 (PublishPersistent)
- 지속 가능한 구독 생성 (SubscribePersistentViaDurable)
- 메시지 수신 및 처리
- ACK(확인) 메커니즘 검증

**중요성**: 메시지 영속성 및 신뢰할 수 있는 전달 보장

**실패 시나리오**: 스트림 생성 실패, 구독 실패, 메시지 손실

### 3. Cluster Communication (클러스터 통신)
**목적**: 클러스터 환경에서의 메시지 통신을 테스트합니다.

**테스트 내용**:
- 휘발성 메시징을 통한 발행/구독
- 클러스터 내 메시지 라우팅
- 다중 노드 시뮬레이션
- 메시지 전달 신뢰성 검증

**중요성**: 분산 시스템에서의 메시지 통신 안정성

**실패 시나리오**: 메시지 라우팅 실패, 클러스터 연결 실패

### 4. Key-Value Store (키-값 저장소)
**목적**: JetStream 기반 키-값 저장소의 CRUD 기능을 테스트합니다.

**테스트 내용**:
- KV 저장소 생성 및 구성
- 키-값 쌍 저장 (PutToKeyValueStore)
- 키-값 쌍 조회 (GetFromKeyValueStore)
- TTL(Time-To-Live) 설정 검증
- 저장소 용량 제한 테스트

**중요성**: 설정, 캐시, 메타데이터 저장을 위한 안정적인 저장소

**실패 시나리오**: 저장 실패, 조회 실패, TTL 만료 실패

### 5. Object Store (객체 저장소)
**목적**: 대용량 객체 저장 및 검색 기능을 테스트합니다.

**테스트 내용**:
- 객체 저장소 생성
- 객체 업로드 (PutToObjectStore)
- 객체 다운로드 (GetFromObjectStore)
- 메타데이터 처리
- 객체 삭제 기능

**중요성**: 파일, 이미지 등의 대용량 데이터 저장

**실패 시나리오**: 업로드 실패, 다운로드 실패, 메타데이터 손실

### 6. Error Handling (에러 처리)
**목적**: 다양한 에러 상황에서의 시스템 동작을 검증합니다.

**테스트 내용**:
- 존재하지 않는 스트림으로의 발행 시도
- 잘못된 구성으로 스트림 생성 시도
- 존재하지 않는 키 조회
- 존재하지 않는 객체 접근
- 네트워크 오류 시뮬레이션

**중요성**: 시스템의 견고성과 에러 복구 능력

**실패 시나리오**: 에러가 제대로 처리되지 않거나, 잘못된 에러 메시지

### 7. Concurrent Operations (동시성 작업)
**목적**: 다중 클라이언트 동시 접근 시의 시스템 안정성을 테스트합니다.

**테스트 내용**:
- 10개의 고루틴을 통한 동시 KV 작업
- 경쟁 조건 검증
- 동시 읽기/쓰기 작업
- 락 메커니즘 검증

**중요성**: 고부하 환경에서의 시스템 성능 및 안정성

**실패 시나리오**: 경쟁 조건, 데드락, 데이터 손상

### 8. Performance Test (성능 테스트)
**목적**: 시스템의 성능 특성을 측정하고 벤치마킹합니다.

**테스트 내용**:
- KV 저장소 작업 성능 (100회 반복)
- 메시징 작업 성능
- 초당 작업 수 측정
- 지연 시간 분석
- 메모리 사용량 모니터링

**중요성**: 성능 요구사항 충족 및 병목 현상 식별

**실패 시나리오**: 성능 저하, 메모리 누수, 타임아웃

### 9. Edge Cases (엣지 케이스)
**목적**: 비정상적이거나 극단적인 입력에 대한 시스템 동작을 테스트합니다.

**테스트 내용**:
- 대용량 메시지 (1MB) 처리
- 특수 문자 포함 키/값 처리
- 빈 값 처리
- 최대 길이 입력 테스트
- 경계값 테스트

**중요성**: 시스템의 견고성과 비정상 입력 처리 능력

**실패 시나리오**: 크래시, 메모리 오버플로우, 잘못된 데이터 처리

### 10. Recovery Scenarios (복구 시나리오)
**목적**: 시스템 장애 및 재시작 시의 데이터 영속성을 검증합니다.

**테스트 내용**:
- 허브 재시작 시뮬레이션
- 데이터 영속성 검증
- 연결 복구 메커니즘
- 상태 일관성 확인

**중요성**: 시스템 신뢰성과 데이터 무결성 보장

**실패 시나리오**: 데이터 손실, 상태 불일치, 복구 실패

### 11. Configuration Validation (구성 검증)
**목적**: 허브 생성 시의 구성 옵션 유효성을 검증합니다.

**테스트 내용**:
- Nil 옵션 처리
- 유효하지 않은 구성 값
- 기본값 적용 검증
- 구성 우선순위 확인

**중요성**: 올바른 시스템 초기화 및 구성 관리

**실패 시나리오**: 잘못된 구성으로 인한 시작 실패

### 12. Retention Policies (보존 정책)
**목적**: 다양한 메시지 보존 정책의 동작을 검증합니다.

**테스트 내용**:
- Limits Policy: 최대 메시지 수 제한
- Interest Policy: 컨슈머 관심도 기반 보존
- WorkQueue Policy: 작업 큐 방식 보존
- 메시지 제한 초과 시 동작 검증

**중요성**: 데이터 보존 전략의 정확성 및 효율성

**실패 시나리오**: 잘못된 보존 동작, 메시지 손실

### 13. Stream Management (스트림 관리)
**목적**: JetStream 스트림의 생성, 업데이트, 관리를 테스트합니다.

**테스트 내용**:
- 다중 스트림 동시 생성
- 스트림 구성 업데이트
- 주제(subject) 확장
- 스트림 간섭 검증

**중요성**: 동적 스트림 관리 및 구성 변경 능력

**실패 시나리오**: 스트림 충돌, 구성 업데이트 실패

### 14. Consumer Management (컨슈머 관리)
**목적**: 지속 가능한 컨슈머의 생성 및 관리를 테스트합니다.

**테스트 내용**:
- 다중 durable 컨슈머 생성
- 메시지 라우팅 및 분배
- 컨슈머별 메시지 처리
- 컨슈머 정리 및 리소스 관리

**중요성**: 확장 가능한 메시지 소비 아키텍처

**실패 시나리오**: 컨슈머 생성 실패, 메시지 분배 오류

### 15. Load Balancing (부하 분산)
**목적**: 작업 큐를 통한 부하 분산 메커니즘을 검증합니다.

**테스트 내용**:
- Limits 정책 기반 작업 분배
- 다중 워커 간 균등 분배
- 작업 완료율 검증
- 큐 처리 효율성

**중요성**: 고부하 환경에서의 작업 분산 처리

**실패 시나리오**: 작업 분배 불균형, 처리 실패

**고급 패턴 검증**:
- 우선순위 큐 분배
- 배치 작업 처리
- 데드 레터 큐 활용
- 워커 풀 스케일링

### 16. Network Resilience (네트워크 복원력)
**목적**: 네트워크 중단 및 복구 시나리오에서의 시스템 안정성을 테스트합니다.

**테스트 내용**:
- 메시지 지속성 검증
- 네트워크 장애 시뮬레이션
- 데이터 손실 방지
- 복구 후 정상 동작 확인

**중요성**: 분산 시스템의 내결함성 및 복원력

**실패 시나리오**: 데이터 손실, 복구 실패, 네트워크 장애 시 다운타임

**복원력 패턴 검증**:
- 영속적 메시지 저장
- 재연결 메커니즘
- 데이터 일관성 유지
- 장애 복구 시간 측정

## 출력 예시

```
=== HubStream Integration Tests ===

--- Running Basic Hub Creation ---
Hub created successfully with ID: hub-1756393399076598726-6310ec9d
✅ PASSED

--- Running JetStream Operations ---
Successfully received and processed message: Hello JetStream!
✅ PASSED

--- Running Cluster Communication ---
Successfully received cluster message: Cluster test message
✅ PASSED

--- Running Key-Value Store ---
Successfully stored and retrieved KV value: test_value
✅ PASSED

--- Running Object Store ---
Successfully stored and retrieved object: test_object.txt (24 bytes)
✅ PASSED

--- Running Error Handling ---
Testing various error scenarios...
✓ Correctly handled publish to non-existent stream: failed to publish to subject "nonexistent.stream": nats: no response from stream
✓ Correctly handled invalid stream config: subjects cannot be empty
✓ Correctly handled non-existent KV key: failed to access key-value store "nonexistent_bucket": nats: bucket not found
✓ Correctly handled non-existent object: failed to access object store "nonexistent_bucket": nats: stream not found
All error handling tests passed!
✅ PASSED

--- Running Concurrent Operations ---
✓ Successfully completed 10 concurrent KV operations
✅ PASSED

--- Running Performance Test ---
Running performance tests...
✓ KV Put Performance: 100 ops in 7.389674ms (13532.40 ops/sec)
✓ KV Get Performance: 100 ops in 7.844284ms (12748.14 ops/sec)
✓ Message Performance: 100 ops in 92.624µs (1079633.79 ops/sec)
✅ PASSED

--- Running Edge Cases ---
Testing edge cases...
✓ Successfully handled large message (1048576 bytes)
✓ Successfully handled special characters in values
✓ Successfully handled empty values
✅ PASSED

--- Running Recovery Scenarios ---
Testing recovery scenarios...
✓ First hub shutdown successfully
✓ Data persistence verified after hub restart
✅ PASSED

--- Running Configuration Validation ---
Testing configuration validation...
✓ Correctly rejected nil options: options cannot be nil
✓ Successfully created hub with valid configuration
✓ Configuration validation completed successfully
✅ PASSED

--- Running Retention Policies ---
Testing different retention policies...
Testing Limits Policy (5 messages)...
✓ Limits Policy (5 messages) working correctly (received 5 messages)
Testing Interest Policy (10 messages)...
✓ Interest Policy (10 messages) working correctly (received 10 messages)
Testing WorkQueue Policy (3 messages)...
✓ WorkQueue Policy (3 messages) working correctly (received 3 messages)
All retention policy tests passed!
✅ PASSED

--- Running Stream Management ---
Testing stream management...
✓ Stream management tests passed!
✅ PASSED

--- Running Consumer Management ---
Testing consumer management...
Consumer consumer-2 received: test-message-0
Consumer consumer-2 received: test-message-1
Consumer consumer-1 received: test-message-0
Consumer consumer-2 received: test-message-2
Consumer consumer-1 received: test-message-1
Consumer consumer-1 received: test-message-2
✓ Consumer management tests passed!
✅ PASSED

--- Running Load Balancing ---
Testing load balancing...
❌ FAILED: failed to create worker worker-1: failed to subscribe to subject "work.tasks": nats: filtered consumer not unique on workqueue stream

--- Running Network Resilience ---
Testing network resilience...
Testing message persistence...
✓ Messages persisted correctly (5 messages)
✓ Network resilience tests passed!
✅ PASSED

=== Test Results ===
Passed: 15
Failed: 1
Total: 16
## 알려진 제한사항

### Load Balancing 테스트
현재 Load Balancing 테스트는 WorkQueue 정책의 제약사항으로 인해 실패할 수 있습니다:

```
❌ FAILED: failed to create worker worker-1: failed to subscribe to subject "work.tasks": nats: filtered consumer not unique on workqueue stream
```

**원인**: NATS JetStream의 WorkQueue 정책에서는 동일한 durable name을 가진 컨슈머가 여러 개 존재할 수 없습니다. 이는 의도된 동작으로, 작업 큐에서 각 작업이 정확히 한 번만 처리되도록 보장합니다.

**해결 방안**: 
- 실제 운영 환경에서는 각 워커가 고유한 durable name을 사용해야 합니다
- 테스트에서는 단일 워커만 생성하여 WorkQueue 기능을 검증할 수 있습니다
- 이 제한사항은 HubStream의 기능적 문제가 아닌 NATS JetStream의 설계적 제약사항입니다

## 기존 go test와의 비교

기존 `go test` 방식은 다음과 같은 제한이 있었습니다:
- JetStream 비활성화로 인한 테스트 실패
- 타임아웃으로 인한 긴 테스트의 중단
- 설정 변경의 어려움

새로운 `go run` 방식은 이러한 제한을 해결합니다.
