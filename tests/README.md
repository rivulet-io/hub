# Hub Testing Framework

이 디렉토리는 Hub 프로젝트의 포괄적인 테스트 스위트를 포함합니다. 각 테스트 파일은 Hub의 다양한 아키텍처 패턴과 배포 시나리오를 검증합니다.

## 📁 테스트 구조 개요

```
tests/
├── main.go          # 테스트 실행 진입점
├── core.go          # 핵심 Hub 기능 테스트
├── cluster.go       # 클러스터링 테스트
├── edge.go          # 엣지 노드 테스트
├── gateway.go       # 게이트웨이 테스트
└── README.md        # 이 문서
```

## 🏗️ 아키텍처 패턴별 테스트

### 1. Core Tests (`core.go`) - 16개 테스트
기본 Hub 기능과 단일 노드 운영을 검증합니다.

| 테스트 이름 | 설명 | 구현 세부사항 |
|------------|------|-------------|
| **Basic Hub Creation** | Hub 인스턴스 생성 및 기본 구성 검증 | `DefaultNodeOptions()`로 옵션 생성, `NewHub()`로 인스턴스화, ID 자동 생성 및 파일 저장 검증 |
| **JetStream Operations** | 지속적 메시징 스트림 기능 | 스트림 생성 → 메시지 발행 → Durable 컨슈머 구독 → ACK 처리 → 메시지 수신 검증 |
| **Cluster Communication** | 내부 통신 메커니즘 | Volatile 메시징을 통한 pub/sub, Fanout 패턴, 메시지 라우팅 및 수신 검증 |
| **Key-Value Store** | KV 저장소 CRUD 작업 | 버킷 생성 → 키-값 저장 → 조회 → 버전 관리 → 삭제 검증 |
| **Object Store** | 객체 저장소 파일 관리 | 버킷 생성 → 파일 업로드 → 메타데이터 설정 → 다운로드 → 삭제 검증 |
| **Error Handling** | 오류 처리 및 복구 | 존재하지 않는 스트림 발행, 빈 subjects 스트림 생성, 존재하지 않는 키 조회 등 오류 시나리오 검증 |
| **Concurrent Operations** | 동시성 처리 | 10개의 고루틴으로 동시 KV 작업 수행, 경쟁 상태 및 데이터 일관성 검증 |
| **Performance Test** | 성능 벤치마킹 | KV put/get 성능 측정, Volatile 메시징 처리량, JetStream 성능 평가 |
| **Edge Cases** | 경계 조건 처리 | 1MB 대용량 메시지, 특수문자 값, 빈 값 처리, 메모리 및 저장소 한계 검증 |
| **Recovery Scenarios** | 장애 복구 시나리오 | Hub 재시작 후 데이터 지속성 검증, 파일 기반 ID 복구, 상태 복원 테스트 |
| **Configuration Validation** | 구성 검증 | nil 옵션 거부, 유효한 최소 구성 허용, 런타임 구성 검증 |
| **Retention Policies** | 데이터 보관 정책 | Limits/Interest/WorkQueue 정책별 메시지 제한 동작 검증 |
| **Stream Management** | 스트림 생명주기 관리 | 다중 스트림 생성, 구성 업데이트, Subject 확장, 스트림 간섭 검증 |
| **Consumer Management** | 컨슈머 관리 | 다중 Durable 컨슈머 생성, 메시지 분배, 컨슈머 격리 및 정리 검증 |
| **Load Balancing** | 부하 분산 | 다중 워커 생성, 메시지 분배 균등성, 처리량 및 응답 시간 검증 |
| **Network Resilience** | 네트워크 복원력 | 메시지 지속성, 네트워크 분할 복구, 데이터 손실 방지 메커니즘 검증 |

### 2. Cluster Tests (`cluster.go`) - 15개 테스트
멀티 노드 클러스터링 및 분산 처리를 검증합니다.

| 테스트 이름 | 설명 | 구현 세부사항 |
|------------|------|-------------|
| **Cluster Formation** | 클러스터 형성 및 노드 결합 | 시드 노드 생성 → 팔로워 노드 연결 → 클러스터 구성 검증 → 노드 간 통신 테스트 |
| **Cluster Basic Communication** | 클러스터 내 노드 간 통신 | 독립 노드 생성 → 메시지 교환 → 네트워크 토폴로지 검증 → 연결성 테스트 |
| **Cluster JetStream Operations** | 분산 JetStream 처리 | 클러스터 스트림 생성 → 분산 메시지 발행 → 복제된 컨슈머 → 데이터 일관성 검증 |
| **Cluster Key-Value Store** | 분산 KV 저장소 | 클러스터 KV 버킷 생성 → 분산 키-값 작업 → 복제 및 동기화 검증 |
| **Cluster Object Store** | 분산 객체 저장소 | 클러스터 객체 버킷 생성 → 분산 파일 저장 → 메타데이터 복제 검증 |
| **Cluster Load Balancing** | 클러스터 부하 분산 | 3개 노드 클러스터 → 큐 기반 작업 분배 → 워커 간 균등 분배 검증 |
| **Cluster Error Handling** | 분산 환경 오류 처리 | 노드 장애 시나리오 → 오류 전파 → 복구 메커니즘 → 클러스터 안정성 검증 |
| **Cluster Node Failure Recovery** | 노드 장애 복구 | 노드 중단 → 데이터 보존 → 재시작 후 복구 → 일관성 검증 |
| **Cluster Concurrent Operations** | 분산 동시성 처리 | 다중 노드 동시 작업 → 경쟁 상태 방지 → 분산 락 메커니즘 검증 |
| **Cluster Performance** | 클러스터 성능 측정 | 분산 KV 성능 → 메시징 처리량 → 네트워크 지연 영향 평가 |
| **Cluster Data Consistency** | 데이터 일관성 보장 | 분산 트랜잭션 → 복제 일관성 → 충돌 해결 메커니즘 검증 |
| **Cluster Network Resilience** | 네트워크 분할 복구 | 네트워크 분할 시뮬레이션 → 파티션 복구 → 데이터 동기화 검증 |
| **Cluster Stream Replication** | 스트림 복제 | 다중 노드 스트림 복제 → 메시지 복제본 → 장애 시 데이터 가용성 검증 |
| **Cluster Consumer Management** | 분산 컨슈머 관리 | 클러스터 컨슈머 그룹 → 워크로드 분배 → 컨슈머 장애 복구 검증 |
| **Cluster Configuration Validation** | 클러스터 구성 검증 | 클러스터 옵션 검증 → 노드 구성 일관성 → 런타임 구성 변경 테스트 |

### 3. Edge Tests (`edge.go`) - 19개 테스트
엣지 컴퓨팅 시나리오와 리프 노드 연결을 검증합니다.

| 테스트 이름 | 설명 | 구현 세부사항 |
|------------|------|-------------|
| **Edge Node Creation** | 엣지 노드 생성 | `DefaultEdgeOptions()`로 엣지 구성 → 리프노드 포트 설정 → 경량화 검증 |
| **Edge to Hub Connection** | 엣지-허브 연결 | 허브 노드 생성 → 엣지 노드 연결 → 리프노드 라우팅 → 메시지 교환 검증 |
| **Edge Node Basic Messaging** | 엣지 노드 메시징 | Fanout/Queue/Request-Reply 패턴 → 엣지 메시지 라우팅 → 허브 통합 검증 |
| **Edge Node JetStream Operations** | 엣지 JetStream 처리 | 엣지 노드 JetStream 비활성화 검증 → 허브 위임 패턴 → 경량 아키텍처 확인 |
| **Edge Node Key-Value Store** | 엣지 KV 저장소 | 엣지 노드 저장소 비활성화 → 허브 저장소 위임 → 분산 저장소 패턴 검증 |
| **Edge Node Object Store** | 엣지 객체 저장소 | 엣지 노드 객체 저장소 비활성화 → 허브 객체 저장소 위임 → 파일 관리 패턴 검증 |
| **Edge Node Message Routing** | 메시지 라우팅 | 엣지-허브 메시지 라우팅 → 주제 기반 라우팅 → 네트워크 홉 최적화 검증 |
| **Edge Node Load Balancing** | 엣지 부하 분산 | 엣지 워커 분배 → 로컬 처리 우선순위 → 허브 오프로드 검증 |
| **Edge Node Failover Recovery** | 엣지 장애 복구 | 엣지 노드 재연결 → 상태 동기화 → 허브 페일오버 검증 |
| **Edge Node Data Synchronization** | 데이터 동기화 | 엣지-허브 데이터 동기화 → 캐시 일관성 → 오프라인 작업 지원 검증 |
| **Edge Node Performance** | 엣지 성능 측정 | 엣지 처리량 → 네트워크 지연 영향 → 로컬 vs 원격 성능 비교 |
| **Edge Node Error Handling** | 엣지 오류 처리 | 네트워크 연결 끊김 → 로컬 오류 처리 → 허브 재연결 복구 검증 |
| **Edge Node Concurrent Operations** | 엣지 동시성 처리 | 엣지 다중 작업 → 리소스 제한 → 우선순위 기반 처리 검증 |
| **Edge Node Network Resilience** | 엣지 네트워크 복원력 | 불안정한 네트워크 → 메시지 큐잉 → 재연결 후 재전송 검증 |
| **Edge Node Configuration Validation** | 엣지 구성 검증 | 엣지 옵션 검증 → 리소스 제한 설정 → 보안 구성 검증 |
| **Edge Node Resource Management** | 엣지 리소스 관리 | 메모리/CPU 제한 → 저장소 할당 → 리소스 모니터링 검증 |
| **Multi-Edge Communication** | 멀티 엣지 통신 | 다중 엣지 노드 → 메시지 브로드캐스트 → 그룹 통신 패턴 검증 |
| **Edge Node Security** | 엣지 보안 | 엣지 인증 → 암호화 통신 → 보안 정책 적용 검증 |
| **Edge Node Monitoring** | 엣지 모니터링 | 엣지 상태 모니터링 → 메트릭 수집 → 허브 중앙 모니터링 검증 |

### 4. Gateway Tests (`gateway.go`) - 19개 테스트
네트워크 간 게이트웨이 연결 및 라우팅을 검증합니다.

| 테스트 이름 | 설명 | 구현 세부사항 |
|------------|------|-------------|
| **Gateway Node Creation** | 게이트웨이 노드 생성 | `DefaultGatewayOptions()`로 게이트웨이 구성 → 게이트웨이 포트 설정 → 라우팅 준비 검증 |
| **Gateway Network Discovery** | 네트워크 발견 | 게이트웨이 노드 생성 → 네트워크 탐색 → 연결 설정 → 토폴로지 구축 검증 |
| **Gateway Inter-Network Routing** | 네트워크 간 라우팅 | 다중 네트워크 생성 → 게이트웨이 연결 → 크로스 네트워크 메시지 라우팅 검증 |
| **Gateway Message Forwarding** | 메시지 전달 | 게이트웨이 메시지 포워딩 → 프로토콜 변환 → 주소 변환 검증 |
| **Gateway Load Balancing** | 게이트웨이 부하 분산 | 다중 게이트웨이 → 트래픽 분배 → 장애 시 자동 전환 검증 |
| **Gateway Failover Recovery** | 게이트웨이 장애 복구 | 게이트웨이 장애 시뮬레이션 → 자동 페일오버 → 서비스 연속성 검증 |
| **Gateway JetStream Operations** | 게이트웨이 JetStream | 게이트웨이 JetStream 지원 → 크로스 네트워크 스트림 복제 → 데이터 동기화 검증 |
| **Gateway Key-Value Store** | 게이트웨이 KV 저장소 | 게이트웨이 KV 버킷 → 크로스 네트워크 데이터 공유 → 복제 전략 검증 |
| **Gateway Object Store** | 게이트웨이 객체 저장소 | 게이트웨이 객체 버킷 → 분산 파일 저장 → 메타데이터 동기화 검증 |
| **Gateway Security and Authentication** | 보안 및 인증 | 게이트웨이 인증 → 암호화 터널 → 보안 정책 적용 검증 |
| **Gateway Performance Monitoring** | 성능 모니터링 | 게이트웨이 처리량 측정 → 지연 시간 모니터링 → 병목 지점 식별 검증 |
| **Gateway Configuration Management** | 구성 관리 | 게이트웨이 구성 → 동적 재구성 → 설정 검증 메커니즘 테스트 |
| **Gateway Network Partitioning** | 네트워크 분할 처리 | 네트워크 분할 시뮬레이션 → 파티션 복구 → 데이터 일관성 유지 검증 |
| **Gateway Concurrent Operations** | 게이트웨이 동시성 | 동시 게이트웨이 작업 → 경쟁 상태 방지 → 확장성 검증 |
| **Gateway Error Handling** | 게이트웨이 오류 처리 | 게이트웨이 오류 시나리오 → 오류 전파 → 복구 전략 검증 |
| **Gateway Resource Management** | 리소스 관리 | 게이트웨이 리소스 할당 → 메모리 관리 → CPU 사용량 최적화 검증 |
| **Multi-Gateway Communication** | 멀티 게이트웨이 통신 | 3개 네트워크 메시 브로드캐스트 → 게이트웨이 메시 라우팅 → 복잡한 토폴로지 검증 |
| **Gateway Cluster Integration** | 클러스터 통합 | 게이트웨이 클러스터링 → 분산 게이트웨이 → 고가용성 구성 검증 |
| **Gateway Monitoring and Metrics** | 모니터링 및 메트릭 | 게이트웨이 상태 모니터링 → 메트릭 수집 → 중앙 집중식 모니터링 검증 |

## 🚀 테스트 실행 방법

### 모든 테스트 실행 (현재 활성화: Gateway)
```bash
cd tests
go run .
```

### 특정 테스트 스위트 실행
`main.go` 파일에서 원하는 테스트 함수를 활성화:

```go
func main() {
    coreTestFunc()        // Core 테스트 실행
    clusterTestFunc()     // Cluster 테스트 실행  
    edgeTestFunc()        // Edge 테스트 실행
    gatewayTestFunc()        // Gateway 테스트 실행 (현재 활성화)
}
```

## 📊 테스트 커버리지

| 테스트 스위트 | 테스트 수 | 주요 검증 영역 |
|-------------|----------|-------------|
| **Core** | 16개 | 기본 기능, 성능, 안정성 |
| **Cluster** | 15개 | 분산 처리, 일관성, 복제 |
| **Edge** | 19개 | 엣지 컴퓨팅, 동기화, 연결성 |
| **Gateway** | 19개 | 네트워크 간 통신, 라우팅 |
| **총합** | **69개** | 전체 아키텍처 패턴 |

## 🔧 테스트 환경 요구사항

- **Go**: 1.19+
- **메모리**: 최소 1GB (동시 노드 실행을 위해)
- **포트**: 4200-4400 범위 (테스트별 격리된 포트 사용)
- **임시 디렉토리**: 테스트 데이터 저장용

## 📋 포트 할당 전략

각 테스트는 포트 충돌을 방지하기 위해 격리된 포트 범위를 사용합니다:

- **Core Tests**: 4200-4220
- **Cluster Tests**: 4224-4250  
- **Edge Tests**: 4260-4270
- **Gateway Tests**: 4280-4400

## 🛠️ 트러블슈팅

### 일반적인 문제들

1. **포트 충돌**: 다른 서비스가 테스트 포트를 사용 중일 경우
   - 해결: 해당 포트를 사용하는 프로세스 종료 또는 포트 변경

2. **타임아웃**: NATS 서버 시작 시간 초과
   - 해결: 시스템 리소스 확인, 방화벽 설정 검토

3. **임시 디렉토리 권한**: 테스트 데이터 저장 실패
   - 해결: 쓰기 권한이 있는 임시 디렉토리 확인

### 성공적인 테스트 실행 예시

```
=== Hub Gateway Integration Tests ===
...
--- Running Gateway Cluster Integration ---
Testing gateway cluster integration...
✓ Seed node created successfully
✓ Second node created successfully
✓ Both nodes are operational
✅ PASSED

=== Gateway Test Results ===
Passed: 19
Failed: 0
Total: 19
```

## 📈 향후 개선 계획

- [ ] 통합 테스트 스위트 (모든 패턴 조합)
- [ ] 성능 벤치마크 자동화
- [ ] 테스트 커버리지 리포팅

---

이 테스트 프레임워크는 Hub 프로젝트의 안정성과 신뢰성을 보장하기 위해 지속적으로 개선됩니다.