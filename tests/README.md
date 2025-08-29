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

| 테스트 이름 | 설명 |
|------------|------|
| Basic Hub Creation | Hub 인스턴스 생성 및 기본 구성 검증 |
| JetStream Operations | 지속적 메시징 스트림 기능 |
| Cluster Communication | 내부 통신 메커니즘 |
| Key-Value Store | KV 저장소 CRUD 작업 |
| Object Store | 객체 저장소 파일 관리 |
| Error Handling | 오류 처리 및 복구 |
| Concurrent Operations | 동시성 처리 |
| Performance Test | 성능 벤치마킹 |
| Edge Cases | 경계 조건 처리 |
| Recovery Scenarios | 장애 복구 시나리오 |
| Configuration Validation | 구성 검증 |
| Retention Policies | 데이터 보관 정책 |
| Stream Management | 스트림 생명주기 관리 |
| Consumer Management | 컨슈머 관리 |
| Load Balancing | 부하 분산 |
| Network Resilience | 네트워크 복원력 |

### 2. Cluster Tests (`cluster.go`) - 15개 테스트
멀티 노드 클러스터링 및 분산 처리를 검증합니다.

| 테스트 이름 | 설명 |
|------------|------|
| Cluster Formation | 클러스터 형성 및 노드 결합 |
| Cluster Basic Communication | 클러스터 내 노드 간 통신 |
| Cluster JetStream Operations | 분산 JetStream 처리 |
| Cluster Key-Value Store | 분산 KV 저장소 |
| Cluster Object Store | 분산 객체 저장소 |
| Cluster Load Balancing | 클러스터 부하 분산 |
| Cluster Error Handling | 분산 환경 오류 처리 |
| Cluster Node Failure Recovery | 노드 장애 복구 |
| Cluster Concurrent Operations | 분산 동시성 처리 |
| Cluster Performance | 클러스터 성능 측정 |
| Cluster Data Consistency | 데이터 일관성 보장 |
| Cluster Network Resilience | 네트워크 분할 복구 |
| Cluster Stream Replication | 스트림 복제 |
| Cluster Consumer Management | 분산 컨슈머 관리 |
| Cluster Configuration Validation | 클러스터 구성 검증 |

### 3. Edge Tests (`edge.go`) - 19개 테스트
엣지 컴퓨팅 시나리오와 리프 노드 연결을 검증합니다.

| 테스트 이름 | 설명 |
|------------|------|
| Edge Node Creation | 엣지 노드 생성 |
| Edge to Hub Connection | 엣지-허브 연결 |
| Edge Node Basic Messaging | 엣지 노드 메시징 |
| Edge Node JetStream Operations | 엣지 JetStream 처리 |
| Edge Node Key-Value Store | 엣지 KV 저장소 |
| Edge Node Object Store | 엣지 객체 저장소 |
| Edge Node Message Routing | 메시지 라우팅 |
| Edge Node Load Balancing | 엣지 부하 분산 |
| Edge Node Failover Recovery | 엣지 장애 복구 |
| Edge Node Data Synchronization | 데이터 동기화 |
| Edge Node Performance | 엣지 성능 측정 |
| Edge Node Error Handling | 엣지 오류 처리 |
| Edge Node Concurrent Operations | 엣지 동시성 처리 |
| Edge Node Network Resilience | 엣지 네트워크 복원력 |
| Edge Node Configuration Validation | 엣지 구성 검증 |
| Edge Node Resource Management | 엣지 리소스 관리 |
| Multi-Edge Communication | 멀티 엣지 통신 |
| Edge Node Security | 엣지 보안 |
| Edge Node Monitoring | 엣지 모니터링 |

### 4. Gateway Tests (`gateway.go`) - 19개 테스트
네트워크 간 게이트웨이 연결 및 라우팅을 검증합니다.

| 테스트 이름 | 설명 |
|------------|------|
| Gateway Node Creation | 게이트웨이 노드 생성 |
| Gateway Network Discovery | 네트워크 발견 |
| Gateway Inter-Network Routing | 네트워크 간 라우팅 |
| Gateway Message Forwarding | 메시지 전달 |
| Gateway Load Balancing | 게이트웨이 부하 분산 |
| Gateway Failover Recovery | 게이트웨이 장애 복구 |
| Gateway JetStream Operations | 게이트웨이 JetStream |
| Gateway Key-Value Store | 게이트웨이 KV 저장소 |
| Gateway Object Store | 게이트웨이 객체 저장소 |
| Gateway Security and Authentication | 보안 및 인증 |
| Gateway Performance Monitoring | 성능 모니터링 |
| Gateway Configuration Management | 구성 관리 |
| Gateway Network Partitioning | 네트워크 분할 처리 |
| Gateway Concurrent Operations | 게이트웨이 동시성 |
| Gateway Error Handling | 게이트웨이 오류 처리 |
| Gateway Resource Management | 리소스 관리 |
| Multi-Gateway Communication | 멀티 게이트웨이 통신 |
| Gateway Cluster Integration | 클러스터 통합 |
| Gateway Monitoring and Metrics | 모니터링 및 메트릭 |

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
- [ ] CI/CD 파이프라인 통합
- [ ] 테스트 커버리지 리포팅
- [ ] 부하 테스트 시나리오 확장

---

이 테스트 프레임워크는 Hub 프로젝트의 안정성과 신뢰성을 보장하기 위해 지속적으로 개선됩니다.