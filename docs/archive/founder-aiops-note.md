# Founder AIOps Note

Historical archive only.
This note is kept for continuity and should not be treated as active runtime authority or default read-path material.

This note was originally kept as a temporary root markdown file during the
early AIOps exploration phase and was moved under `docs/` during repo cleanup.

## Original Note

코딩에이전트 세팅 여러가지 삽질해보고 얻은 깨달음  
최근들어 가장 신기한 변화라고 느껴집니다. 원래 어떤 서비스를 만들면, 서비스를 운영하는 사람이 유저들의 데이터를 보고 서비스를 개선합니다. 더욱 최적화된 서비스를 제공하기 위해 노력합니다. 사용자 모두가 점차 평균적으로 높은 만족도의 UX를 경험합니다.  
하지만 AI 서비스들은 그 한계가 명확한 듯 합니다. LLM은 모델 자체 특수한 목적보단 범용적인 목적을 위해 설계되어 만들어진 모델입니다. 모든 사용자의 edge case에 맞게 맞춰줄 수가 없습니다. 그러다보니 누군가는 성공적으로 세팅하여 10x, 20x의 효율을 내지만 누군가는 기본 설정으로 끝없이 뒤쳐져 갑니다. 뒤쳐지고 있는지도 모르는 상태에서 말이죠.  

저 또한 매일 나오는 소식들에 뒤쳐질까 두려움이 생기는 것이 사실입니다. 그래서 늘 팔로업하기 위해 노력하고 최대한 많은 부분들을 자동화하려고 노력합니다.  
최근에 가장 흥미로웠던 프로젝트는 aiops 였습니다. Codex를 사용하는 패턴을 분석하고 그에 맞는 mcp, hook, skills, custom instruction을 제안, 자동으로 백그라운드에서 적용해주는 에이전트였습니다.  
저는 그냥 평상시대로 코딩하다보면 환경이 자동으로 점점 진화하는 합니다. 처음에는 어설펐는데 시간이 지나며 저도 몰랐던 제 패턴들에 맞는 custom instruction들이 쌓여 가면서 완전히 다른 성능의 모델을 사용하는 것처럼 체감되는 것 같네요.  
꽤나 오랫동안 이런저런 세팅도 해보기 위해 삽질해봤던 저에게는 놀라운 경험이었습니다. 앞으로 발전되는 모델들에서 개인이 하는 커스터마이징이 만들 더욱 더 큰 격차가 체감되었습니다.  

자동화 시스템을 구축하고 나서야 제가 추가해놓은 mcp들이 엄청나게 무의미한 토큰들을 소모했고, 비용이나 시간 지연만 만들었다는 것들을 알고 불필요한 것들을 모두 정리했습니다. 사실 이것도 제가 아니라 aiops 에이전트가 정리해줬습니다.  
토큰을 소모하는 속도가 크게 줄었고, 이제는 Rate limit에서 훨씬 자유로워졌네요. 제가 자주 시키는 일들, 모델이 자주 틀리는 일들을 자동으로 instruction에 기록해주니까 작업이 매끄러워지기도 했고, 삽질 시간이 줄어드니까 확실히 빠릿빠릿하네요.  
aiops라는 이름의 이 프로젝트는 최근에 했던 프로젝트들 중에 가장 성공적이었고, 앞으로도 계속 업그레이드 해 나가고 싶네요.  
미래에는 각자의 환경에 모델이나 세팅을 평가할 수 있는 personal harness가 갖춰지지 않을까 예상해봅니다.
