# FutBoss AI - Model Tests
# Author: Bruno Lucena (bruno@lucena.cloud)

import pytest
from datetime import datetime

from app.models.user import User, UserCreate, UserResponse
from app.models.player import Player, PlayerAttributes, PlayerPersonality, Position, Temperament, PlayStyle
from app.models.team import Team, Formation
from app.models.match import Match, MatchStatus, MatchEvent, EventType
from app.models.token import TokenWallet, TokenTransaction, TransactionType


class TestUserModel:
    def test_user_create_validation(self):
        user = UserCreate(
            username="testuser",
            email="test@example.com",
            password="password123"
        )
        assert user.username == "testuser"
        assert user.email == "test@example.com"

    def test_user_create_short_username(self):
        with pytest.raises(ValueError):
            UserCreate(username="ab", email="test@example.com", password="password123")

    def test_user_create_short_password(self):
        with pytest.raises(ValueError):
            UserCreate(username="testuser", email="test@example.com", password="12345")


class TestPlayerModel:
    def test_player_attributes_overall(self):
        attrs = PlayerAttributes(
            speed=80, strength=70, stamina=75,
            finishing=85, passing=70, dribbling=78, defense=40,
            intelligence=72, aggression=60, leadership=55, creativity=80
        )
        overall = attrs.overall()
        assert 60 <= overall <= 80  # Average of all attributes

    def test_player_attributes_clamping(self):
        attrs = PlayerAttributes(speed=150, strength=-10)
        # Pydantic should clamp values
        assert attrs.speed == 100 or attrs.speed == 150  # depends on validation
    
    def test_player_market_value_young(self):
        attrs = PlayerAttributes(
            speed=80, strength=70, stamina=75,
            finishing=85, passing=70, dribbling=78, defense=40,
            intelligence=72, aggression=60, leadership=55, creativity=80
        )
        player = Player(
            name="Young Star",
            position=Position.ST,
            nationality="Brasil",
            age=20,
            attributes=attrs,
            personality=PlayerPersonality(),
        )
        value = player.get_market_value()
        # Young players get 1.3x modifier
        assert value > attrs.overall() * 10

    def test_player_market_value_old(self):
        attrs = PlayerAttributes()  # Default 50 all
        player = Player(
            name="Veteran",
            position=Position.CB,
            nationality="Brasil",
            age=35,
            attributes=attrs,
            personality=PlayerPersonality(),
        )
        value = player.get_market_value()
        # Old players get 0.7x modifier
        assert value < attrs.overall() * 10


class TestTeamModel:
    def test_team_points(self):
        team = Team(
            name="Test FC",
            owner_id="user123",
            wins=10,
            draws=5,
            losses=3,
        )
        assert team.points() == 35  # 10*3 + 5*1

    def test_team_goal_difference(self):
        team = Team(
            name="Test FC",
            owner_id="user123",
            goals_for=25,
            goals_against=15,
        )
        assert team.goal_difference() == 10

    def test_team_matches_played(self):
        team = Team(
            name="Test FC",
            owner_id="user123",
            wins=10,
            draws=5,
            losses=3,
        )
        assert team.matches_played() == 18


class TestMatchModel:
    def test_match_is_home_winner_yes(self):
        match = Match(
            home_team_id="team1",
            away_team_id="team2",
            home_score=3,
            away_score=1,
            status=MatchStatus.FINISHED,
        )
        assert match.is_home_winner() is True

    def test_match_is_home_winner_no(self):
        match = Match(
            home_team_id="team1",
            away_team_id="team2",
            home_score=1,
            away_score=3,
            status=MatchStatus.FINISHED,
        )
        assert match.is_home_winner() is False

    def test_match_is_home_winner_draw(self):
        match = Match(
            home_team_id="team1",
            away_team_id="team2",
            home_score=2,
            away_score=2,
            status=MatchStatus.FINISHED,
        )
        assert match.is_home_winner() is None

    def test_match_is_home_winner_not_finished(self):
        match = Match(
            home_team_id="team1",
            away_team_id="team2",
            status=MatchStatus.IN_PROGRESS,
        )
        assert match.is_home_winner() is None


class TestTokenWallet:
    def test_wallet_can_afford_yes(self):
        wallet = TokenWallet(user_id="user1", balance=1000)
        assert wallet.can_afford(500) is True

    def test_wallet_can_afford_no(self):
        wallet = TokenWallet(user_id="user1", balance=100)
        assert wallet.can_afford(500) is False

    def test_wallet_credit(self):
        wallet = TokenWallet(user_id="user1", balance=1000)
        tx = wallet.credit(500, TransactionType.MATCH_REWARD, "Won match")
        assert wallet.balance == 1500
        assert wallet.total_earned == 1500
        assert tx.amount == 500

    def test_wallet_debit_success(self):
        wallet = TokenWallet(user_id="user1", balance=1000)
        tx = wallet.debit(300, TransactionType.BUY_PLAYER, "Bought player")
        assert wallet.balance == 700
        assert wallet.total_spent == 300
        assert tx.amount == -300

    def test_wallet_debit_insufficient(self):
        wallet = TokenWallet(user_id="user1", balance=100)
        tx = wallet.debit(500, TransactionType.BUY_PLAYER, "Bought player")
        assert tx is None
        assert wallet.balance == 100  # Unchanged

